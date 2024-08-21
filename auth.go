package fengchaogo

import (
	"context"
	"fmt"
	"time"
)

const ExpiresTime = 1700

// authManager 授权控制器
type authManager struct {
	accessToken string
	expiresAt   time.Time
}

// tokenResponse token响应
type tokenResponse struct {
	Status int    `json:"status"`
	Token  string `json:"token"`
	Msg    string `json:"msg"`
}

// getAuthToken 获取token
func (f *FengChao) getAuthToken() (string, error) {

	if f.auth == nil || time.Since(f.auth.expiresAt) > time.Duration(ExpiresTime)*time.Second {
		err := f.refreshToken()
		if err != nil {
			return "", err
		}
	}

	return f.auth.accessToken, nil
}

// refreshToken 刷新token
func (f *FengChao) refreshToken() error {
	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(BasicRequestTimeout)*time.Second)
	defer cancel()
	resp, err := f.client.R().
		SetContext(ctx).
		SetDebug(false).
		SetLogger(nil).
		SetQueryParam("api_key", f.ApiKey).
		SetQueryParam("secret_key", f.SecretKey).
		SetResult(&tokenResponse{}).
		Get("/token")

	if err != nil {
		return fmt.Errorf("get auth token client error: %v", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("get auth token response error: %v", resp)
	}

	if resp.Result().(*tokenResponse).Status != 200 {
		return fmt.Errorf("get auth token error: %v", resp.Result().(*tokenResponse).Msg)
	}

	f.auth = &authManager{
		accessToken: resp.Result().(*tokenResponse).Token,
		expiresAt:   time.Now().Add(time.Duration(ExpiresTime) * time.Second),
	}

	return nil
}
