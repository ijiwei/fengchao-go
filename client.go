package fengchaogo

import (
	"sync"

	"github.com/go-resty/resty/v2"
)

const BasicRequestTimeout int = 3

// FengChaoOptions 配置
type FengChao struct {
	// ApiKey fengchao api key
	ApiKey string
	// SecretKey fengchao secret key
	SecretKey string
	// BaseUrl api url
	BaseUrl string

	// client http client
	client *resty.Client

	// authToken 认证令牌
	auth *authManager

	// availableModels 可用模型
	availableModels *modelsManager

	sync.Mutex
}

func NewFengChao(apiKey string, secretKey string, baseUrl string) *FengChao {
	fengChao := &FengChao{
		ApiKey:    apiKey,
		SecretKey: secretKey,
		BaseUrl:   baseUrl,
	}

	client := resty.New().
		SetBaseURL(fengChao.BaseUrl).
		SetDebug(false)
	fengChao.client = client
	return fengChao
}

func (f *FengChao) SetDebug(debug bool) *FengChao {
	f.client.SetDebug(debug)
	return f
}
