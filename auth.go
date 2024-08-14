package fengchaogo

import (
	"context"
	"fmt"
	"time"
)

const ExpiresTime = 1700

type AuthToken struct {
	accessToken string
	expiresIn   int64
}

type AuthResult struct {
	Status int    `json:"status"`
	Token  string `json:"token"`
	Msg    string `json:"msg"`
}

// getAuthToken 获取token
func (f *FengChao) getAuthToken(ctx context.Context) (string, error) {
	if f.authToken == nil || f.authToken.expiresIn < time.Now().Unix() {
		err := f.refreshToken(ctx)
		if err != nil {
			return "", err
		}
	}

	return f.authToken.accessToken, nil
}

// refreshToken 刷新token
func (f *FengChao) refreshToken(ctx context.Context) error {
	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer cancel()
	resp, err := f.client.R().
		SetContext(ctx).
		SetQueryParam("api_key", f.ApiKey).
		SetQueryParam("secret_key", f.SecretKey).
		SetResult(&AuthResult{}).
		Get("/token")

	if err != nil {
		return fmt.Errorf("get auth token client error: %v", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("get auth token response error: %v", resp)
	}

	if resp.Result().(*AuthResult).Status != 200 {
		return fmt.Errorf("get auth token error: %v", resp.Result().(*AuthResult).Msg)
	}

	f.authToken = &AuthToken{
		accessToken: resp.Result().(*AuthResult).Token,
		expiresIn:   time.Now().Unix() + int64(ExpiresTime),
	}

	return nil
}
