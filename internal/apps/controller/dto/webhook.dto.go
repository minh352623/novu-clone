package dto

import (
	"time"

	"CONVERDA/internal/apps/domain/model/entity"

	"github.com/google/uuid"
)

// Webhook DTOs
type CreateWebhookRequest struct {
	EnvironmentID       uuid.UUID `json:"environment_id" binding:"required"`
	URL                 string    `json:"url" binding:"required,url"`
	Events              []string  `json:"events"`
	Description         *string   `json:"description"`
	MaxRetries          *int      `json:"max_retries,omitempty"`
	RetryBackoffSeconds *int      `json:"retry_backoff_seconds,omitempty"`
	RetryTimeoutSeconds *int      `json:"retry_timeout_seconds,omitempty"`
}

type UpdateWebhookRequest struct {
	URL                 *string  `json:"url"`
	Events              []string `json:"events"`
	Description         *string  `json:"description"`
	IsActive            *bool    `json:"is_active"`
	MaxRetries          *int     `json:"max_retries,omitempty"`
	RetryBackoffSeconds *int     `json:"retry_backoff_seconds,omitempty"`
	RetryTimeoutSeconds *int     `json:"retry_timeout_seconds,omitempty"`
}

type WebhookResponse struct {
	ID                  uuid.UUID `json:"id"`
	TenantID            uuid.UUID `json:"tenant_id"`
	AppID               uuid.UUID `json:"app_id"`
	EnvironmentID       uuid.UUID `json:"environment_id"`
	URL                 string    `json:"url"`
	Secret              string    `json:"secret"`
	Description         *string   `json:"description,omitempty"`
	Events              []string  `json:"events"`
	IsActive            bool      `json:"is_active"`
	MaxRetries          int       `json:"max_retries"`
	RetryBackoffSeconds int       `json:"retry_backoff_seconds"`
	RetryTimeoutSeconds int       `json:"retry_timeout_seconds"`
	CreatedAt           time.Time `json:"created_at"`
}

func ToWebhookResponse(w *entity.Webhook) *WebhookResponse {
	if w == nil {
		return nil
	}
	return &WebhookResponse{
		ID:                  w.ID,
		TenantID:            w.TenantID,
		AppID:               w.AppID,
		EnvironmentID:       w.EnvironmentID,
		URL:                 w.URL,
		Secret:              w.Secret,
		Description:         w.Description,
		Events:              w.Events,
		IsActive:            w.IsActive,
		MaxRetries:          w.MaxRetries,
		RetryBackoffSeconds: w.RetryBackoffSeconds,
		RetryTimeoutSeconds: w.RetryTimeoutSeconds,
		CreatedAt:           w.CreatedAt,
	}
}

func ToWebhookResponseList(webhooks []*entity.Webhook) []*WebhookResponse {
	list := make([]*WebhookResponse, 0, len(webhooks))
	for _, w := range webhooks {
		list = append(list, ToWebhookResponse(w))
	}
	return list
}

// Provider DTOs
type CreateProviderRequest struct {
	EnvironmentID uuid.UUID              `json:"environment_id" binding:"required"`
	ProviderType  string                 `json:"provider_type" binding:"required"`
	ProviderName  string                 `json:"provider_name" binding:"required"`
	Configuration map[string]interface{} `json:"configuration"`
}

type UpdateProviderRequest struct {
	ProviderName  *string                `json:"provider_name"`
	Configuration map[string]interface{} `json:"configuration"`
	IsActive      *bool                  `json:"is_active"`
}

type ProviderResponse struct {
	ID            uuid.UUID              `json:"id"`
	TenantID      uuid.UUID              `json:"tenant_id"`
	AppID         uuid.UUID              `json:"app_id"`
	EnvironmentID uuid.UUID              `json:"environment_id"`
	ProviderType  string                 `json:"provider_type"`
	ProviderName  string                 `json:"provider_name"`
	IsActive      bool                   `json:"is_active"`
	Configuration map[string]interface{} `json:"configuration"`
	CreatedAt     time.Time              `json:"created_at"`
}

func ToProviderResponse(p *entity.Provider) *ProviderResponse {
	if p == nil {
		return nil
	}
	return &ProviderResponse{
		ID:            p.ID,
		TenantID:      p.TenantID,
		AppID:         p.AppID,
		EnvironmentID: p.EnvironmentID,
		ProviderType:  p.ProviderType,
		ProviderName:  p.ProviderName,
		IsActive:      p.IsActive,
		Configuration: p.Configuration,
		CreatedAt:     p.CreatedAt,
	}
}

func ToProviderResponseList(providers []*entity.Provider) []*ProviderResponse {
	list := make([]*ProviderResponse, 0, len(providers))
	for _, p := range providers {
		list = append(list, ToProviderResponse(p))
	}
	return list
}
