package fengchaogo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
)

// 消息的角色, user and assistant and system
const (
	// RoleUser  用户消息
	RoleUser = "user"
	// RoleAssistant  机器人消息
	RoleAssistant = "assistant"
	// RoleSystem  系统消息
	RoleSystem = "system"
)

// Message 消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`

	template *template.Template
	buffer   *bytes.Buffer
}

var _ Prompt = (*Message)(nil)

// excute 模板变量的应用
func (m *Message) execute(vairables map[string]interface{}) error {
	if m.template == nil {
		return fmt.Errorf("message template is empty, can not execute")
	}
	if m.buffer == nil {
		m.buffer = &bytes.Buffer{}
	}

	if err := m.template.Execute(m.buffer, vairables); err != nil {
		return err
	}
	m.Content = m.buffer.String()
	m.buffer.Reset()

	return nil
}

// checkRole 检查角色
func (m *Message) checkRole() error {
	if m.Role != RoleUser && m.Role != RoleAssistant && m.Role != RoleSystem {
		return fmt.Errorf("message role is invalid")
	}
	return nil
}

// Render 消息模板渲染为Json
func (m *Message) Render(vairables map[string]interface{}) ([]byte, error) {
	if err := m.checkRole(); err != nil {
		return nil, err
	}
	if m.template != nil {
		err := m.execute(vairables)
		if err != nil {
			return nil, err
		}
	}

	return json.Marshal(m)
}

// RenderMessages 消息模板渲染为消息切片
func (m *Message) RenderMessages(vairables map[string]interface{}) ([]*Message, error) {
	if err := m.checkRole(); err != nil {
		return nil, err
	}

	if m.template != nil {
		err := m.execute(vairables)
		if err != nil {
			return nil, err
		}
	}

	return []*Message{m}, nil
}

// lazyMessage 预渲染消息
type lazyMessage func() (*Message, error)

var _ Prompt = (lazyMessage)(nil)

// Render 消息模板渲染为Json
func (message lazyMessage) Render(vairables map[string]interface{}) ([]byte, error) {
	msg, err := message()
	if err != nil {
		return nil, err
	}

	return msg.Render(vairables)
}

// RenderMessages 消息模板渲染为消息切片
func (message lazyMessage) RenderMessages(vairables map[string]interface{}) ([]*Message, error) {
	msg, err := message()
	if err != nil {
		return nil, err
	}
	return msg.RenderMessages(vairables)
}

// NewMessage 生成消息（这个消息是预渲染的消息）
func NewMessage(role string, messageStr string) lazyMessage {

	return func() (*Message, error) {
		template, err := template.New("").Parse(messageStr)
		if err != nil {
			return nil, fmt.Errorf("parse message template error: %v", err)
		}
		return &Message{Role: role, template: template}, nil
	}
}
