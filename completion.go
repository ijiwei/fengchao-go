package fengchaogo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
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
	Stop []string
	// Timeout 超时时间
	Timeout int
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

// WithModel 设置模型
func WithModel(model string) Option[ChatCompletion] {
	return func(option *ChatCompletion) {
		option.Model = model
	}
}

// WithTemperature 设置模型参数
func WithTemperature(temperature float64) Option[ChatCompletion] {
	return func(option *ChatCompletion) {
		option.Temperature = temperature
	}
}

// WithTopP 设置模型参数
func WithTopP(topP float64) Option[ChatCompletion] {
	return func(option *ChatCompletion) {
		option.TopP = topP
	}
}

// WithDoSample 设置是否开启采样
func WithDoSample(doSample bool) Option[ChatCompletion] {
	return func(option *ChatCompletion) {
		option.DoSample = doSample
	}
}

// WithMaxTokens 设置最大长度
func WithMaxTokens(maxTokens int) Option[ChatCompletion] {
	return func(option *ChatCompletion) {
		option.MaxTokens = maxTokens
	}
}

// WithStop 设置停用词
func WithStop(stop []string) Option[ChatCompletion] {
	return func(option *ChatCompletion) {
		option.Stop = stop
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout int) Option[ChatCompletion] {
	return func(option *ChatCompletion) {
		option.Timeout = timeout
	}
}

// WithQuery 设置问题
func WithQuery(query string) Option[ChatCompletion] {
	return func(option *ChatCompletion) {
		option.Query = query
	}
}

// withPredefinedPrompts 设置预定义的prompt提示工程
func WithPredefinedPrompts(predefinedPrompts string) Option[ChatCompletion] {
	return func(option *ChatCompletion) {
		option.PredefinedPrompts = predefinedPrompts
	}
}

// WithSystem 设置系统消息
func WithSystem(system string) Option[ChatCompletion] {
	return func(option *ChatCompletion) {
		option.System = system
	}
}

// WithIsSensitive 设置是否开启敏感词
func WithIsSensitive(isSensitive bool) Option[ChatCompletion] {
	return func(option *ChatCompletion) {
		option.IsSensitive = isSensitive
	}
}

// WithVariables 设置变量
func WithParams(variables any) Option[ChatCompletion] {
	t := reflect.TypeOf(variables)
	// 先判断是否为指针类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	v := reflect.ValueOf(variables)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	var mapVars = make(map[string]interface{})

	// 首先判断是否为结构体
	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			currentField := v.Field(i)
			if !currentField.CanInterface() {
				continue
			}
			mapVars[t.Field(i).Name] = currentField.Interface()
		}
	}

	// 再给一次机会,判断是否为map
	if t.Kind() == reflect.Map {
		// 判断是否为map[string]any
		// 如果不是就不行
		if t.Key().Kind() == reflect.String {
			// 遍历
			for _, key := range v.MapKeys() {
				// 获取键和对应的值
				k := key.Interface()
				if v.MapIndex(key).CanInterface() {
					v := v.MapIndex(key).Interface()
					mapVars[k.(string)] = v
				}
			}
		}
	}

	return func(option *ChatCompletion) {
		option.variables = mapVars
	}
}

// WithRequestID 设置请求ID
func WithRequestID(requestID string) Option[ChatCompletion] {
	return func(option *ChatCompletion) {
		option.RequestID = requestID
	}
}

// Apply 应用配置
func (option *ChatCompletion) Apply(helpers ...Option[ChatCompletion]) {
	for _, helper := range helpers {
		helper(option)
	}
}

// RenderMessages 渲染消息列表
func (option *ChatCompletion) LoadPromptTemplates(prompt Prompt) ([]*Message, error) {
	var messages []*Message
	originalMessages := make([]*Message, len(messages))
	if prompt == nil {
		return originalMessages, nil
	}
	messages, err := prompt.RenderMessages(option.variables)
	if err != nil {
		return nil, fmt.Errorf("render message template with error[%v]", err)
	}
	// 去掉第一个是系统消息
	copy(originalMessages, messages)
	if messages[0].Role == RoleSystem {
		option.System = messages[0].Content
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

	option.Query = query
	option.History = messages
	return originalMessages, nil
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
	Detail []ChatCompletionErrorDetail `json:"detail"`
}

// String 聊天错误信息
func (c *ChatCompletionError) String() string {
	if c.Detail == nil || len(c.Detail) == 0 {
		return "unknown error"
	}
	return c.Detail[0].Msg
}

// ChatCompletionErrorDetail 聊天错误详情
type ChatCompletionErrorDetail struct {
	Msg string `json:"msg"`
}

// VerifyError 验证错误,实现了StreamAble
func chatCompletionErrorHandler(ccr ChatCompletionResult) error {
	if ccr.Status != 200 {
		return fmt.Errorf("chat completion failed: [%d]%s", ccr.Status, ccr.Msg)
	}

	return nil
}

// ChatCompletion 聊天
func (f *FengChao) ChatCompletion(ctx context.Context, prompt Prompt, chatCompletionOption ...Option[ChatCompletion]) (*ChatCompletionResult, error) {
	ChatCompletionOption := defaultChatCompletionOption()
	ChatCompletionOption.Apply(chatCompletionOption...)
	originalMessages, err := ChatCompletionOption.LoadPromptTemplates(prompt)
	if err != nil {
		return nil, fmt.Errorf("fail to load prompt template cause: %s", err)
	}

	AvailableModles := f.GetAvailableModels()
	if AvailableModles == nil {
		return nil, fmt.Errorf("available model is empty, please check service")
	}
	var found *Model
	currentModel := ChatCompletionOption.Model
	for _, model := range AvailableModles {
		if currentModel == model.ID {
			found = &model
		}
	}
	if found == nil {
		return nil, fmt.Errorf("unsupport model (%s)", currentModel)
	}

	var uri = "/chat/"
	if found.Channel == "本地模型" {
		uri = "/local_chat/"
	}

	token, err := f.getAuthToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("fail to auth cause: %s", err)
	}
	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, time.Duration(ChatCompletionOption.Timeout)*time.Second)
	defer cancel()
	resp, err := f.client.R().
		SetContext(ctx).
		SetBody(ChatCompletionOption).
		SetHeaderMultiValues(map[string][]string{
			"Content-Type":  {"application/json"},
			"Authorization": {token},
		}).
		SetError(&ChatCompletionError{}).
		SetResult(&ChatCompletionResult{}).
		Post(uri)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("request timeout")
		}
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("chat completion error: %s", resp.Error().(*ChatCompletionError).String())
	}
	if resp.Result().(*ChatCompletionResult).Status != 200 {
		return nil, fmt.Errorf("error[%d]: %v", resp.Result().(*ChatCompletionResult).Status, resp.Result().(*ChatCompletionResult).Msg)
	}
	complettionResult := resp.Result().(*ChatCompletionResult)
	complettionResult.History = append(originalMessages, &Message{
		Role:    RoleAssistant,
		Content: complettionResult.String(),
	})
	return complettionResult, nil
}

// QuickCompletion 使用预定义prompt, 快速生成文本
func (f *FengChao) QuickCompletion(ctx context.Context, chatCompletionOption ...Option[ChatCompletion]) (*ChatCompletionResult, error) {
	ChatCompletionOption := defaultChatCompletionOption()
	ChatCompletionOption.Apply(chatCompletionOption...)
	if ChatCompletionOption.PredefinedPrompts == "" || ChatCompletionOption.Query == "" {
		return nil, fmt.Errorf("prompt or query is empty")
	}
	AvailableModles := f.GetAvailableModels()
	if AvailableModles == nil {
		return nil, fmt.Errorf("available model is empty, please check service")
	}
	var found *Model
	currentModel := ChatCompletionOption.Model
	for _, model := range AvailableModles {
		if currentModel == model.ID {
			found = &model
		}
	}
	if found == nil {
		return nil, fmt.Errorf("unsupport model (%s)", currentModel)
	}

	var uri = "/chat/"
	if found.Channel == "本地模型" {
		uri = "/local_chat/"
	}

	token, err := f.getAuthToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("fail to auth cause: %s", err)
	}
	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, time.Duration(ChatCompletionOption.Timeout)*time.Second)
	defer cancel()
	resp, err := f.client.R().
		SetContext(ctx).
		SetBody(ChatCompletionOption).
		SetHeaderMultiValues(map[string][]string{
			"Content-Type":  {"application/json"},
			"Authorization": {token},
		}).
		SetError(&ChatCompletionError{}).
		SetResult(&ChatCompletionResult{}).
		Post(uri)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("request timeout")
		}
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("chat completion error: %s", resp.Error().(*ChatCompletionError).String())
	}
	if resp.Result().(*ChatCompletionResult).Status != 200 {
		return nil, fmt.Errorf("error[%d]: %v", resp.Result().(*ChatCompletionResult).Status, resp.Result().(*ChatCompletionResult).Msg)
	}
	complettionResult := resp.Result().(*ChatCompletionResult)
	return complettionResult, nil
}

// String 聊天结果
func (r *ChatCompletionResult) String() string {
	if r.Choices == nil || len(r.Choices) == 0 {
		return ""
	}
	return r.Choices[0].Message.Content
}

// GetHistoryPrompts 获取历史消息（Prompt）
func (r *ChatCompletionResult) GetHistoryPrompts() *PromptTemplate {
	if r.History == nil || len(r.History) == 0 {
		return nil
	}
	prompts := make([]Prompt, 0)
	for _, m := range r.History {
		prompts = append(prompts, m)
	}
	return NewPromptTemplate(prompts...)
}
