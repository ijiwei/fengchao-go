package fengchaogo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
)

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`

	template *template.Template
	buffer   bytes.Buffer
}

func (m *Message) execute(vairables map[string]interface{}) error {

	if err := m.template.Execute(&m.buffer, vairables); err != nil {
		return err
	}
	m.Content = m.buffer.String()

	if m.Content == "" {
		return fmt.Errorf("message content is empty")
	}
	if m.Role != RoleUser && m.Role != RoleAssistant && m.Role != RoleSystem {
		return fmt.Errorf("message role is invalid")
	}
	return nil
}

func (m *Message) Render(vairables map[string]interface{}) ([]byte, error) {
	if m.template == nil {
		return json.Marshal(m)
	}
	err := m.execute(vairables)
	if err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

func (m *Message) RenderMessages(vairables map[string]interface{}) ([]Message, error) {
	if m.template == nil {
		return []Message{*m}, nil
	}
	err := m.execute(vairables)
	if err != nil {
		return nil, err
	}
	return []Message{*m}, nil
}

type lazyMessage func() (*Message, error)

var _ Prompt = (lazyMessage)(nil)

func (message lazyMessage) Render(vairables map[string]interface{}) ([]byte, error) {
	msg, err := message()
	if err != nil {
		return nil, err
	}

	return msg.Render(vairables)
}

func (message lazyMessage) RenderMessages(vairables map[string]interface{}) ([]Message, error) {
	msg, err := message()
	if err != nil {
		return nil, err
	}
	return msg.RenderMessages(vairables)
}

func NewMessage(role string, messageStr string) lazyMessage {
	if role != RoleUser && role != RoleAssistant && role != RoleSystem {
		role = RoleUser
	}
	return func() (*Message, error) {
		template, err := template.New("").Parse(messageStr)
		if err != nil {
			return nil, fmt.Errorf("parse message template error: %v", err)
		}
		return &Message{Role: role, template: template}, nil
	}
}
