package fengchaogo

import (
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

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
	authToken *AuthToken

	// Logger 日志
	logger *logrus.Logger
}

func NewFengChao(apiKey string, secretKey string, baseUrl string) *FengChao {
	fengChao := &FengChao{
		ApiKey:    apiKey,
		SecretKey: secretKey,
		BaseUrl:   baseUrl,
	}

	client := resty.New().
		SetBaseURL(fengChao.BaseUrl).
		SetLogger(nil)
	fengChao.client = client
	fengChao.logger = logrus.StandardLogger()
	return fengChao
}

func (f *FengChao) SetLogger(logger *logrus.Logger) *FengChao {
	f.logger = logger
	return f
}
