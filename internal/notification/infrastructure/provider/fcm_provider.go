package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"CONVERDA/global"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

type FcmConfig struct {
	ServiceAccountJSON string `json:"service_account_json"`
}

type fcmProvider struct {
	client *messaging.Client
}

func NewFcmProvider(configBytes []byte) (PushProvider, error) {
	var cfg FcmConfig
	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse fcm config: %w", err)
	}

	// For MOCK/Placeholder purpose if JSON is empty or invalid
	if cfg.ServiceAccountJSON == "" || cfg.ServiceAccountJSON == "{}" {
		global.Logger.Warn("FCM: using MOCK client (no service account JSON provided)")
		return &mockFcmProvider{}, nil
	}

	opt := option.WithCredentialsJSON([]byte(cfg.ServiceAccountJSON))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %v", err)
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting messaging client: %v", err)
	}

	return &fcmProvider{client: client}, nil
}

func (p *fcmProvider) Send(ctx context.Context, to, subject, body string, data map[string]string) error {
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: subject,
			Body:  body,
		},
		Data:  data,
		Token: to,
	}

	response, err := p.client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send fcm message: %w", err)
	}

	global.Logger.Info("FCM: successfully sent push notification", zap.String("response", response))
	return nil
}

// mockFcmProvider for development without real credentials
type mockFcmProvider struct{}

func (p *mockFcmProvider) Send(ctx context.Context, to, subject, body string, data map[string]string) error {
	global.Logger.Info("FCM-MOCK: sending push", zap.String("to", to))
	global.Logger.Info("FCM-MOCK: push content", zap.String("title", subject), zap.String("body", body))
	return nil
}
