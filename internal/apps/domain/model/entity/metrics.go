package entity

import (
	"time"

	"github.com/google/uuid"
)

// Message direction constants
const (
	DirectionInbound  = "inbound"
	DirectionOutbound = "outbound"
)

type UsageMetric struct {
	ID            uuid.UUID `json:"id"`
	AppID         uuid.UUID `json:"app_id"`
	EnvironmentID uuid.UUID `json:"environment_id"`
	ProviderType  string    `json:"provider_type"` // email, sms, push, etc.
	Direction     string    `json:"direction"`     // inbound, outbound
	StatusCode    int       `json:"status_code"`
	Timestamp     time.Time `json:"timestamp"`
}

func NewUsageMetric(appID, envID uuid.UUID, providerType, direction string, statusCode int) *UsageMetric {
	if direction == "" {
		direction = DirectionOutbound
	}
	return &UsageMetric{
		ID:            uuid.New(),
		AppID:         appID,
		EnvironmentID: envID,
		ProviderType:  providerType,
		Direction:     direction,
		StatusCode:    statusCode,
		Timestamp:     time.Now(),
	}
}

// --- Aggregation result types ---

// DetailedMetricsSummary provides a full breakdown of usage for an app.
type DetailedMetricsSummary struct {
	TotalMessages    int64                     `json:"total_messages"`
	InboundMessages  int64                     `json:"inbound_messages"`
	OutboundMessages int64                     `json:"outbound_messages"`
	SuccessCount     int64                     `json:"success_count"`
	FailureCount     int64                     `json:"failure_count"`
	ByProvider       map[string]ProviderMetric `json:"by_provider"`
}

// ProviderMetric is a per-provider breakdown.
type ProviderMetric struct {
	Total   int64 `json:"total"`
	Success int64 `json:"success"`
	Failed  int64 `json:"failed"`
}

// DailyMetric is a single day's aggregated data.
type DailyMetric struct {
	Date     string `json:"date"` // "2026-02-17"
	Inbound  int64  `json:"inbound"`
	Outbound int64  `json:"outbound"`
	Success  int64  `json:"success"`
	Failed   int64  `json:"failed"`
}

// EnvironmentMetric is a per-environment breakdown.
type EnvironmentMetric struct {
	EnvironmentID uuid.UUID `json:"environment_id"`
	Total         int64     `json:"total"`
	Inbound       int64     `json:"inbound"`
	Outbound      int64     `json:"outbound"`
	Success       int64     `json:"success"`
	Failed        int64     `json:"failed"`
}
