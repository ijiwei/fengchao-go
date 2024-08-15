package fengchaogo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
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
	History []*Message `json:"history"`
	// Query 问题
	Query string `json:"query"`
	// System 系统消息
	System string `json:"system"`

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

// WithHistory 设置历史消息
func WithHistory(history []*Message) Option[ChatCompletion] {
	return func(option *ChatCompletion) {
		option.History = history
	}
}

// WithQuery 设置问题
func WithQuery(query string) Option[ChatCompletion] {
	return func(option *ChatCompletion) {
		option.Query = query
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

	// 首先判断是否为结构体
	if t.Kind() == reflect.Struct {
		var mapVars = make(map[string]interface{})
		for i := 0; i < t.NumField(); i++ {
			current := reflect.ValueOf(variables)
			if current.Kind() == reflect.Ptr {
				current = current.Elem()
			}
			if !current.Field(i).CanInterface() {
				fmt.Printf("该字段不可获取 %s\n", t.Field(i).Name)
				continue
			}
			mapVars[t.Field(i).Name] = current.Field(i).Interface()
		}
		return func(option *ChatCompletion) {
			option.variables = mapVars
		}
	}

	// 再给一次机会，判断是否为map
	if t.Kind() == reflect.Map {
		// 判断是否为map[string]any
		// 如果不是就不行
		if t.Key().Kind() == reflect.String {
			val := reflect.ValueOf(variables)
			if val.Kind() == reflect.Ptr {
				val = val.Elem()
			}
			var mapVars = make(map[string]interface{})
			// 遍历
			for _, key := range val.MapKeys() {
				// 获取键和对应的值
				k := key.Interface()
				v := val.MapIndex(key).Interface()

				mapVars[k.(string)] = v
			}
			return func(option *ChatCompletion) {
				option.variables = mapVars
			}
		}
	}
	return func(option *ChatCompletion) {
		option.variables = nil
	}
}

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
	messages, err := prompt.RenderMessages(option.variables)
	if err != nil {
		return nil, fmt.Errorf("render message template with error[%v]", err)
	}
	originalMessages := make([]*Message, len(messages))
	// 去掉第一个是系统消息
	copy(originalMessages, messages)
	if messages[0].Role == RoleSystem {
		option.System = messages[0].Content
		messages = messages[1:]
	}
	if len(messages) == 0 {
		return nil, errors.New("user messages is empty")
	}
	query := messages[len(messages)-1].Content
	messages = messages[:len(messages)-1]

	option.Query = query
	option.History = messages
	return originalMessages, nil
}

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
		SetResult(&ChatCompletionResult{}).
		Post(uri)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("request timeout")
		}
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("response error")
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

func (r *ChatCompletionResult) String() string {
	if r.Choices == nil || len(r.Choices) == 0 {
		return ""
	}
	return r.Choices[0].Message.Content
}

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
