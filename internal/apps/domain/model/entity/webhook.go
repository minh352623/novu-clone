package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrWebhookNotFound  = errors.New("webhook not found")
	ErrProviderNotFound = errors.New("provider not found")
)

type Webhook struct {
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
	UpdatedAt           time.Time `json:"updated_at"`
}

func NewWebhook(tenantID, appID, environmentID uuid.UUID, url string, events []string, description *string) (*Webhook, error) {
	if url == "" {
		return nil, errors.New("webhook url is required")
	}
	// Generate secret
	secret := "whsec_" + uuid.New().String()

	return &Webhook{
		ID:            uuid.New(),
		TenantID:      tenantID,
		AppID:         appID,
		EnvironmentID: environmentID,
		URL:           url,
		Secret:        secret,
		Description:   description,
		Events:        events,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
}
