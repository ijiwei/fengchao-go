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

type ModelsResponse struct {
	Data []Model `json:"data"`
}

type ModelsManager struct {
	model      []Model
	lastUpdate time.Time
}

func (f *FengChao) GetModels() []Model {
	if f.AvailableModels == nil || time.Since(f.AvailableModels.lastUpdate) > 24*time.Hour {
		err := f.loadModels(context.Background())
		if err != nil {
			return nil
		}
	}
	return f.AvailableModels.model
}

func (f *FengChao) loadModels(ctx context.Context) error {
	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()
	resp, err := f.client.R().
		SetContext(ctx).
		SetResult(&ModelsResponse{}).
		Get("/models/")

	if err != nil {
		return fmt.Errorf("get models error: %v", err)
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("response error")
	}

	f.AvailableModels = &ModelsManager{
		model:      resp.Result().(*ModelsResponse).Data,
		lastUpdate: time.Now(),
	}
	return nil
}
