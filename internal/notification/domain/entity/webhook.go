package entity

import (
	"time"

	"github.com/google/uuid"
)

type Webhook struct {
	ID                  uuid.UUID
	TenantID            uuid.UUID
	EnvironmentID       uuid.UUID
	URL                 string
	Secret              string // HMAC Secret
	Events              []string
	Description         *string
	IsActive            bool
	MaxRetries          int
	RetryBackoffSeconds int
	RetryTimeoutSeconds int
	AppID               uuid.UUID
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type WebhookLog struct {
	ID             uuid.UUID
	WebhookID      uuid.UUID
	URL            string
	EventType      string
	RequestPayload map[string]interface{}
	ResponseCode   int
	ResponseBody   *string
	DurationMs     int64
	Status         string // 'pending', 'success', 'failed'
	RetryCount     int
	MaxRetries     int
	NextRetryAt    *time.Time
	TenantID       uuid.UUID
	AppID          uuid.UUID
	CreatedAt      time.Time
}

const (
	WebhookLogStatusPending = "pending"
	WebhookLogStatusSuccess = "success"
	WebhookLogStatusFailed  = "failed"
)
