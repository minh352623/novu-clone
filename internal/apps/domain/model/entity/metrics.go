package entity

import (
	"time"

	"github.com/google/uuid"
)

type UsageMetric struct {
	ID            uuid.UUID `json:"id"`
	AppID         uuid.UUID `json:"app_id"`
	EnvironmentID uuid.UUID `json:"environment_id"`
	ProviderType  string    `json:"provider_type"` // email, sms, push, etc.
	StatusCode    int       `json:"status_code"`
	Timestamp     time.Time `json:"timestamp"`
}

func NewUsageMetric(appID, envID uuid.UUID, providerType string, statusCode int) *UsageMetric {
	return &UsageMetric{
		ID:            uuid.New(),
		AppID:         appID,
		EnvironmentID: envID,
		ProviderType:  providerType,
		StatusCode:    statusCode,
		Timestamp:     time.Now(),
	}
}
