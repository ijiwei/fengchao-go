package fengchaogo

import (
	"context"
	"fmt"
	"time"
)

type Model struct {
	ID             string   `json:"id"`
	OwnedBy        string   `json:"owned_by"`
	MaxInputToken  int      `json:"max_input_token"`
	MaxOutputToken int      `json:"max_output_token"`
	InPrice        float64  `json:"in_price"`
	OutPrice       float64  `json:"out_price"`
	Unit           string   `json:"unit"`
	Modes          []string `json:"mode"`
	Channel        string   `json:"channel"`
	Created        string   `json:"created"`
}

type modelsResponse struct {
	Data []Model `json:"data"`
}

type modelsManager struct {
	models     []Model
	lastUpdate time.Time
}

func (f *FengChao) GetAvailableModels() []Model {
	if f.availableModels == nil || time.Since(f.availableModels.lastUpdate) > 24*time.Hour {
		err := f.loadModels(context.Background())
		if err != nil {
			return nil
		}
	}
	return f.availableModels.models
}

func (f *FengChao) loadModels(ctx context.Context) error {
	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, time.Duration(BasicRequestTimeout)*time.Second)
	defer cancel()
	resp, err := f.client.R().
		SetContext(ctx).
		SetResult(&modelsResponse{}).
		Get("/models/")

	if err != nil {
		return fmt.Errorf("get models error: %v", err)
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("response error")
	}

	f.availableModels = &modelsManager{
		models:     resp.Result().(*modelsResponse).Data,
		lastUpdate: time.Now(),
	}
	return nil
}
