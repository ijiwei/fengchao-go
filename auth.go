package fengchaogo

import (
	"context"
	"fmt"
	"time"
)

const ExpiresTime = 1700

type AuthToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

type AuthResult struct {
	Status int    `json:"status"`
	Token  string `json:"token"`
	Msg    string `json:"msg"`
}

// getAuthToken 获取token
func (f *FengChao) getAuthToken(ctx context.Context) (string, error) {
	if f.authToken != nil && f.authToken.ExpiresIn > time.Now().Unix() {
		return f.authToken.AccessToken, nil
	}

	err := f.refreshToken(ctx)
	if err != nil {
		return "", err
	}

	return f.authToken.AccessToken, nil
}

// refreshToken 刷新token
func (f *FengChao) refreshToken(ctx context.Context) error {
	resp, err := f.client.R().
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

	if resp.Result().(*AuthResult).Status != 0 {
		return fmt.Errorf("get auth token error: %v", resp.Result().(*AuthResult).Msg)
	}

	f.authToken = &AuthToken{
		AccessToken: resp.Result().(*AuthResult).Token,
		ExpiresIn:   time.Now().Unix() + int64(ExpiresTime),
	}

	return nil
}
