package fengchaogo

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"net/http"
)

// ChatCompletionStream 流式聊天
func (f *FengChao) ChatCompletionStream(ctx context.Context, prompt Prompt, chatCompletionOption ...Option[ChatCompletion]) (*JsonStreamReader[ChatCompletionResult], error) {
	ChatCompletionParams := NewChatCompletion(chatCompletionOption...)
	ChatCompletionParams.Mode = StreamMode

	_, err := ChatCompletionParams.LoadPromptTemplates(prompt)
	if err != nil {
		return nil, fmt.Errorf("fail to load prompt template cause: %s", err)
	}

	token, err := f.getAuthToken()
	if err != nil {
		return nil, fmt.Errorf("fail to auth cause: %s", err)
	}

	resp, err := f.client.R().
		SetContext(ctx).
		SetBody(ChatCompletionParams).
		SetHeaderMultiValues(map[string][]string{
			"Content-Type":  {"application/json"},
			"Authorization": {token},
		}).
		SetDoNotParseResponse(true).
		Post("/chat/")

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
func (f *FengChao) ChatCompletionStreamSimple(ctx context.Context, prompt Prompt, chatCompletionOption ...Option[ChatCompletion]) (iter.Seq[ChatCompletionResult], error) {
	reader, err := f.ChatCompletionStream(ctx, prompt, chatCompletionOption...)
	return reader.Stream(), err
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
