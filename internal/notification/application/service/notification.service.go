package service

import (
	"context"
)

type SendRequest struct {
	TenantID      string                 `json:"tenant_id"`
	EnvironmentID string                 `json:"environment_id"`
	TemplateCode  string                 `json:"template_code"`
	Recipient     string                 `json:"recipient"`
	Channel       string                 `json:"channel"`
	Data          map[string]interface{} `json:"data"`
	Language      string                 `json:"language"`
}

type SendResponse struct {
	NotificationID string `json:"notification_id"`
	Status         string `json:"status"`
}

type NotificationService interface {
	Send(ctx context.Context, req SendRequest) (*SendResponse, error)
}
