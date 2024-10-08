package fengchaogo

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ChatCompletionStream 流式聊天
func (f *FengChao) ChatCompletionStream(ctx context.Context, prompt Prompt, chatCompletionOption ...Option[ChatCompletion]) (*JsonStreamReader[ChatCompletionResult], error) {
	ChatCompletionOption := defaultChatCompletionOption()
	ChatCompletionOption.Mode = StreamMode
	ChatCompletionOption.Apply(chatCompletionOption...)

	_, err := ChatCompletionOption.LoadPromptTemplates(prompt)
	if err != nil {
		return nil, fmt.Errorf("fail to load prompt template cause: %s", err)
	}

	model := f.getModel(ChatCompletionOption.Model)
	if model == nil {
		return nil, fmt.Errorf("unsupport model (%s)", ChatCompletionOption.Model)
	}

	var uri = "/chat/"
	if model.Channel == "本地模型" {
		uri = "/local_chat/"
	}

	token, err := f.getAuthToken()
	if err != nil {
		return nil, fmt.Errorf("fail to auth cause: %s", err)
	}

	resp, err := f.client.R().
		SetContext(ctx).
		SetBody(ChatCompletionOption).
		SetHeaderMultiValues(map[string][]string{
			"Content-Type":  {"application/json"},
			"Authorization": {token},
		}).
		SetDoNotParseResponse(true).
		Post(uri)

	if err != nil {
		return nil, fmt.Errorf("fail to post request cause: %s", err)
	}

	if resp.StatusCode() != 200 {
		return nil, handleErrorResponse(resp.RawResponse)
	}

	reader := &JsonStreamReader[ChatCompletionResult]{
		reader:       bufio.NewReader(resp.RawResponse.Body),
		resp:         resp.RawResponse,
		errorHandler: chatCompletionErrorHandler,
	}

	return reader, err
}

// ChatCompletionStreamSimple 流式聊天
func (f *FengChao) ChatCompletionStreamSimple(ctx context.Context, prompt Prompt, chatCompletionOption ...Option[ChatCompletion]) (<-chan *ChatCompletionResult, error) {
	reader, err := f.ChatCompletionStream(ctx, prompt, chatCompletionOption...)
	return reader.Stream(ctx), err
}

// handleErrorResponse 处理错误
func handleErrorResponse(resp *http.Response) error {
	buffer := bufio.NewReader(resp.Body)
	data, err := buffer.ReadString('\n')
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return fmt.Errorf("response unknown error: %s", data)
		}
	}
	var chatCompletionError ChatCompletionError
	err = json.Unmarshal([]byte(data), &chatCompletionError)
	if err != nil {
		return fmt.Errorf("response error: %s", data)
	}
	return fmt.Errorf("chat completion error: %s", chatCompletionError.String())
}
