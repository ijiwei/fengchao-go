package fengchaogo

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Model 模型
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

// modelsResponse 获取模型方法的响应
type modelsResponse struct {
	Data []Model `json:"data"`
}

// modelsManager 模型管理器
type modelsManager struct {
	models     []Model
	lastUpdate time.Time
}

// GetAvailableModels 获取可用模型
func (f *FengChao) GetAvailableModels() []Model {
	if f.availableModels == nil || time.Since(f.availableModels.lastUpdate) > 24*time.Hour {
		err := f.loadModels()
		if err != nil {
			return nil
		}
	}
	return f.availableModels.models
}

// loadModels 加载模型
func (f *FengChao) loadModels() error {
	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(BasicRequestTimeout)*time.Second)
	defer cancel()
	resp, err := f.client.R().
		SetContext(ctx).
		SetDebug(false).
		SetLogger(nil).
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

// getModel 获取模型
func (f *FengChao) getModel(name string) (m *Model) {
	if strings.Contains(name, ",") {
		name = name[:strings.Index(name, ",")]
	}
	// 只校验第一个模型
	for _, model := range f.GetAvailableModels() {
		if name == model.ID {
			m = &model
		}
	}
	return m
}
