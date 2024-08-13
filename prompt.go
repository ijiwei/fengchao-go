package fengchaogo

import (
	"encoding/json"
	"fmt"
)

type Prompt interface {
	Render(vairables map[string]interface{}) ([]byte, error)
}

var _ Prompt = (*PromptTemplate)(nil)

type PromptTemplate struct {
	Messages []Message
	Prompts  []Prompt

	HumanFriendly bool
}

func NewPromptTemplate(p ...Prompt) *PromptTemplate {
	return &PromptTemplate{
		Prompts: p,
	}
}

func (m *PromptTemplate) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Messages)
}

// Render 渲染 Prompt
func (m *PromptTemplate) Render(vairables map[string]interface{}) ([]byte, error) {
	defer func() {
		m.Messages = nil
	}()
	if err := m.renderMessage(vairables); err != nil {
		return nil, err
	}
	messages := m.Messages
	if !m.HumanFriendly {
		return json.Marshal(messages)
	}
	return json.MarshalIndent(messages, "", "  ")
}

// renderMessage 渲染
func (m *PromptTemplate) renderMessage(vairables map[string]interface{}) error {
	if len(m.Prompts) == 0 {
		return fmt.Errorf("prompt template is empty")
	}

	for _, item := range m.Prompts {
		switch item := item.(type) {
		case *PromptTemplate:
			promptTemplate := item
			if err := promptTemplate.renderMessage(vairables); err != nil {
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
