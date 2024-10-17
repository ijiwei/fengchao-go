package fengchaogo

import (
	"context"
	"fmt"
	"sync"
)

// BatchSize 批量请求最大数量
const BatchMaxSize = 5

// BatchChatCompletionArgs 批量请求参数
type BatchChatCompletionArgs struct {
	Prompt Prompt
	Params []Option[ChatCompletion]
}

// BatchChatCompletionBuilder 批量请求创建器
type BatchChatCompletionBuilder struct {
	Args []*BatchChatCompletionArgs
	size int
}

// NewBatchChatCompletionBuilder 创建
func NewBatchChatCompletionBuilder() *BatchChatCompletionBuilder {
	return &BatchChatCompletionBuilder{
		Args: make([]*BatchChatCompletionArgs, 0, BatchMaxSize),
	}
}

// Add 添加
func (bccb *BatchChatCompletionBuilder) Add(prompt Prompt, params ...Option[ChatCompletion]) (*BatchChatCompletionArgs, error) {
	if bccb.size >= BatchMaxSize {
		return nil, fmt.Errorf("batch size exceeded")
	}
	arg := &BatchChatCompletionArgs{
		Prompt: prompt,
		Params: params,
	}
	bccb.Args = append(bccb.Args, arg)

	bccb.size++
	return arg, nil
}

// BatchChatCompletion 批量请求
func (f *FengChao) BatchChatCompletion(ctx context.Context, bccb *BatchChatCompletionBuilder) (map[*BatchChatCompletionArgs]*ChatCompletionResult, map[*BatchChatCompletionArgs]error, bool) {
	completions := make(map[*BatchChatCompletionArgs]*ChatCompletionResult, len(bccb.Args))
	errors := make(map[*BatchChatCompletionArgs]error, len(bccb.Args))
	wg := new(sync.WaitGroup)
	var commplete = true
	for _, arg := range bccb.Args {

		wg.Add(1)
		go func(cca *BatchChatCompletionArgs) {
			prompt := cca.Prompt
			params := cca.Params
			completion, err := f.ChatCompletion(ctx, prompt, params...)
			if err != nil {
				wg.Done()
				errors[cca] = err
				commplete = false
				return
			}
			completions[cca] = completion
			wg.Done()
		}(arg)
	}
	wg.Wait()

	return completions, errors, commplete
}
