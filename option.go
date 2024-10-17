package fengchaogo

import "reflect"

// OptionHelper 配置
type Option[T any] func(option *T)

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
