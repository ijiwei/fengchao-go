package fengchaogo

import (
	"encoding/json"
	"fmt"
)

// Prompt 接口, 暴露一个渲染的能力
type Prompt interface {
	// Render 渲染，这个用来展示
	Render(vairables map[string]interface{}) ([]byte, error)
	// RenderMessages 渲染消息列表，对应的渲染方法是 Render，这个提供给用户自定义使用
	RenderMessages(vairables map[string]interface{}) ([]Message, error)
}

// PromptTemplate 模板
type PromptTemplate struct {
	Messages []Message
	Prompts  []Prompt

	HumanFriendly bool
}

var _ Prompt = (*PromptTemplate)(nil)

// NewPromptTemplate 创建 PromptTemplate
func NewPromptTemplate(p ...Prompt) *PromptTemplate {
	return &PromptTemplate{
		Prompts: p,
	}
}

// MarshalJSON 渲染
func (m *PromptTemplate) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Messages)
}

// Render 渲染 Prompt
func (m *PromptTemplate) Render(vairables map[string]interface{}) ([]byte, error) {
	messages, err := m.RenderMessages(vairables)
	if err != nil {
		return nil, err
	}
	if !m.HumanFriendly {
		return json.Marshal(messages)
	}
	return json.MarshalIndent(messages, "", "  ")
}

// RenderMessages 渲染消息列表
func (m *PromptTemplate) RenderMessages(vairables map[string]interface{}) ([]Message, error) {
	defer func() {
		m.Messages = nil
	}()
	if err := m.execute(vairables); err != nil {
		return nil, err
	}
	messages := m.Messages
	return messages, nil
}

// renderMessage 渲染
func (m *PromptTemplate) execute(vairables map[string]interface{}) error {
	if len(m.Prompts) == 0 {
		return fmt.Errorf("prompt template is empty")
	}

	for _, item := range m.Prompts {
		switch item := item.(type) {
		case *PromptTemplate:
			promptTemplate := item
			if err := promptTemplate.execute(vairables); err != nil {
				return err
			}
			m.Messages = promptTemplate.Messages
		case *Message:
			message := item
			m.Messages = append(m.Messages, *message)
		case lazyMessage:
			message, err := item()
			if err != nil {
				return fmt.Errorf("load lazy message error cause %v", err)
			}
			err = message.execute(vairables)
			if err != nil {
				return fmt.Errorf("execute lazy message error cause %v", err)
			}
			m.Messages = append(m.Messages, *message)
		default:
			continue
		}
	}

	return nil
}
