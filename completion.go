package fengchaogo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	InvokeMode = "invoke"
	StreamMode = "stream"
)

type ChatCompletion struct {
	// RequestID 请求ID
	RequestID string `json:"request_id"`
	// Model 模型
	Model string `json:"model"`
	// Temperature 模型参数
	Temperature float64 `json:"temperature,omitempty"`
	// TopP 模型参数
	TopP float64 `json:"top_p,omitempty"`
	// DoSample 是否开启采样
	DoSample bool `json:"do_sample"`
	// IsSensitive 是否开启敏感词
	IsSensitive bool `json:"is_sensitive"`
	// MaxTokens 最大长度
	MaxTokens int `json:"max_tokens,omitempty"`
	// Stop 停用词
	History []*Message `json:"history,omitempty"`
	// Query 问题
	Query string `json:"query"`
	// System 系统消息
	System string `json:"system"`
	// Mode 是否流式返回
	Mode string `json:"mode,omitempty"`
	// PredefinedPrompts 预定义的prompt提示工程
	PredefinedPrompts string `json:"prompt,omitempty"`

	// Variables 变量
	variables map[string]interface{}

	// Stop 停用词
	Stop []string `json:"-"`
	// Timeout 超时时间
	Timeout int `json:"-"`
}

// DefaultChatCompletionOption 默认配置, 可以覆盖
var DefaultChatCompletionOption = &ChatCompletion{
	Model:       "ERNIE-Bot-4",
	Stop:        []string{},
	MaxTokens:   2000,
	Timeout:     60,
	IsSensitive: false,
}

// 获取默认配置
func defaultChatCompletionOption() *ChatCompletion {
	return DefaultChatCompletionOption.Clone()
}

// clone 拷贝
func (cc *ChatCompletion) Clone() *ChatCompletion {
	clone := *cc
	clone.RequestID = uuid.New().String() // 确保生成新的 RequestID
	return &clone
}

// String 转字符串
func (cc *ChatCompletion) String() string {
	return fmt.Sprintf("RequestID: %s, Model: %s, Query: %s, Temperature: %f, TopP: %f", cc.RequestID, cc.Model, cc.Query, cc.Temperature, cc.TopP)
}

// Apply 应用配置
func (cc *ChatCompletion) Apply(helpers ...Option[ChatCompletion]) {
	for _, helper := range helpers {
		helper(cc)
	}
}

// RenderMessages 渲染消息列表
func (cc *ChatCompletion) LoadPromptTemplates(prompt Prompt) ([]*Message, error) {
	var messages []*Message
	originalMessages := make([]*Message, len(messages))
	if prompt == nil {
		return originalMessages, nil
	}
	messages, err := prompt.RenderMessages(cc.variables)
	if err != nil {
		return nil, fmt.Errorf("render message template with error[%v]", err)
	}
	// 去掉第一个是系统消息
	copy(originalMessages, messages)
	if messages[0].Role == RoleSystem {
		cc.System = messages[0].Content
		messages = messages[1:]
	}
	if len(messages) == 0 {
		return nil, errors.New("user messages is empty")
	}
	if messages[len(messages)-1].Role != RoleUser {
		return nil, errors.New("last message must be user role message")
	}
	query := messages[len(messages)-1].Content
	messages = messages[:len(messages)-1]

	cc.Query = query
	cc.History = messages
	return originalMessages, nil
}

// 使用配置创建一个ChatCompletion参数
func NewChatCompletion(helpers ...Option[ChatCompletion]) *ChatCompletion {
	ChatCompletionOption := defaultChatCompletionOption()
	ChatCompletionOption.Apply(helpers...)
	return ChatCompletionOption
}

// ChatCompletionResult 聊天结果
type ChatCompletionResult struct {
	RequestID string `json:"request_id"`
	Object    string `json:"object"`
	Created   string `json:"created"`
	Choices   []struct {
		Index        int     `json:"index"`
		Role         string  `json:"role"`
		FinishReason string  `json:"finish_reason"`
		Message      Message `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Msg     string `json:"msg"`
	Status  int    `json:"status"`
	History []*Message
}

// ChatCompletionError 聊天错误
type ChatCompletionError struct {
	Detail string `json:"detail"`
}

// String 聊天错误信息
func (cce *ChatCompletionError) String() string {
	return cce.Detail
}

// VerifyError 验证错误,实现了StreamAble
func chatCompletionErrorHandler(ccr ChatCompletionResult) error {
	if ccr.Status != 200 {
		return fmt.Errorf("chat completion failed: [%d]%s", ccr.Status, ccr.Msg)
	}

	return nil
}

func (ccr *ChatCompletionResult) HandleError() error {
	return chatCompletionErrorHandler(*ccr)
}

// String 获取结果的正文内容字符串
func (r *ChatCompletionResult) String() string {
	if r.Choices == nil {
		return ""
	}
	if len(r.Choices) == 0 {
		return ""
	}
	return r.Choices[0].Message.Content
}

// GetHistoryPrompts 获取历史消息（Prompt）
func (r *ChatCompletionResult) GetHistoryPrompts() *PromptTemplate {
	if r.History == nil {
		return nil
	}
	if len(r.History) == 0 {
		return nil
	}
	prompts := make([]Prompt, 0)
	for _, m := range r.History {
		prompts = append(prompts, m)
	}
	return NewPromptTemplate(prompts...)
}

// ChatCompletion 聊天
func (f *FengChao) ChatCompletion(ctx context.Context, prompt Prompt, chatCompletionOption ...Option[ChatCompletion]) (*ChatCompletionResult, error) {
	ChatCompletionParams := NewChatCompletion(chatCompletionOption...)

	originalMessages, err := ChatCompletionParams.LoadPromptTemplates(prompt)
	if err != nil {
		return nil, fmt.Errorf("fail to load prompt template cause: %s", err)
	}

	token, err := f.getAuthToken()
	if err != nil {
		return nil, fmt.Errorf("auth failed, %s", err)
	}

	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, time.Duration(ChatCompletionParams.Timeout)*time.Second)
	defer cancel()
	resp, err := f.client.R().
		SetContext(ctx).
		SetBody(ChatCompletionParams).
		SetHeaderMultiValues(map[string][]string{
			"Content-Type":  {"application/json"},
			"Authorization": {token},
		}).
		SetError(&ChatCompletionError{}).
		SetResult(&ChatCompletionResult{}).
		Post("/chat/")

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("request timeout")
		}
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("chat completion error: %s", resp.Error().(*ChatCompletionError).String())
	}

	complettionResult := resp.Result().(*ChatCompletionResult)

	if err := complettionResult.HandleError(); err != nil {
		return complettionResult, err
	}

	complettionResult.History = append(originalMessages, &Message{
		Role:    RoleAssistant,
		Content: complettionResult.String(),
	})

	return complettionResult, nil
}

// QuickCompletion 使用预定义prompt, 快速生成文本
func (f *FengChao) QuickCompletion(ctx context.Context, chatCompletionOption ...Option[ChatCompletion]) (*ChatCompletionResult, error) {
	ChatCompletionParams := NewChatCompletion(chatCompletionOption...)

	if ChatCompletionParams.PredefinedPrompts == "" || ChatCompletionParams.Query == "" {
		return nil, fmt.Errorf("prompt or query is empty")
	}

	token, err := f.getAuthToken()
	if err != nil {
		return nil, fmt.Errorf("fail to auth cause: %s", err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(ChatCompletionParams.Timeout)*time.Second)
	defer cancel()
	resp, err := f.client.R().
		SetContext(ctx).
		SetBody(ChatCompletionParams).
		SetHeaderMultiValues(map[string][]string{
			"Content-Type":  {"application/json"},
			"Authorization": {token},
		}).
		SetError(&ChatCompletionError{}).
		SetResult(&ChatCompletionResult{}).
		Post("/chat/")

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("request timeout")
		}
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("chat completion error: %s", resp.Error().(*ChatCompletionError).String())
	}

	complettionResult := resp.Result().(*ChatCompletionResult)

	if err := complettionResult.HandleError(); err != nil {
		return complettionResult, err
	}

	return complettionResult, nil
}
