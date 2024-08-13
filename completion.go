package fengchaogo

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
	DoSample bool `json:"do_sample,omitempty"`
	// MaxTokens 最大长度
	MaxTokens int `json:"max_tokens,omitempty"`

	Variables map[string]interface{}
	// Stop 停用词
	Stop []string
	// MaxRetry 最大重试次数
	MaxRetry int
	// Timeout 超时时间
	Timeout int
}

// 获取默认配置
func NewChatCompletionOption() *ChatCompletion {
	return &ChatCompletion{
		Model:    "gpt-4o",
		Stop:     []string{},
		MaxRetry: 2,
		Timeout:  60,
	}
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

// WithMaxRetry 设置最大重试次数
func WithMaxRetry(maxRetry int) Option[ChatCompletion] {
	return func(option *ChatCompletion) {
		option.MaxRetry = maxRetry
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout int) Option[ChatCompletion] {
	return func(option *ChatCompletion) {
		option.Timeout = timeout
	}
}

// Apply 应用配置
func (option *ChatCompletion) Apply(helpers ...Option[ChatCompletion]) {
	for _, helper := range helpers {
		helper(option)
	}
}

type ChatCompletionResult struct {
	RequestID string `json:"request_id"`
	Object    string `json:"object"`
	Created   int    `json:"created"`
	Choices   []struct {
		Index        int    `json:"index"`
		Role         string `json:"role"`
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// func (f *FengChao) ChatCompletion(ctx context.Context, chatCompletionOption ...Option[ChatCompletion]) (*ChatCompletionResult, error) {
// 	ChatCompletionOption := NewChatCompletionOption()
// 	ChatCompletionOption.Apply(chatCompletionOption...)

// 	resp, err := f.client.R().
// 		SetBody(&ChatCompletionOption).Post("/chat/")
// }
