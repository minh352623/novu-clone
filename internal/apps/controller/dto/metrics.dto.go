package dto

import (
	"CONVERDA/internal/apps/domain/model/entity"

	"github.com/google/uuid"
)

// --- Detailed Metrics Response ---

type DetailedMetricsResponse struct {
	TotalMessages    int64                        `json:"total_messages"`
	InboundMessages  int64                        `json:"inbound_messages"`
	OutboundMessages int64                        `json:"outbound_messages"`
	SuccessCount     int64                        `json:"success_count"`
	FailureCount     int64                        `json:"failure_count"`
	SuccessRate      float64                      `json:"success_rate"` // percentage
	ByProvider       map[string]ProviderMetricDTO `json:"by_provider"`
}

type ProviderMetricDTO struct {
	Total   int64 `json:"total"`
	Success int64 `json:"success"`
	Failed  int64 `json:"failed"`
}

func ToDetailedMetricsResponse(s *entity.DetailedMetricsSummary) *DetailedMetricsResponse {
	var successRate float64
	if s.TotalMessages > 0 {
		successRate = float64(s.SuccessCount) / float64(s.TotalMessages) * 100
	}

	byProvider := make(map[string]ProviderMetricDTO)
	for k, v := range s.ByProvider {
		byProvider[k] = ProviderMetricDTO{
			Total:   v.Total,
			Success: v.Success,
			Failed:  v.Failed,
		}
	}

	return &DetailedMetricsResponse{
		TotalMessages:    s.TotalMessages,
		InboundMessages:  s.InboundMessages,
		OutboundMessages: s.OutboundMessages,
		SuccessCount:     s.SuccessCount,
		FailureCount:     s.FailureCount,
		SuccessRate:      successRate,
		ByProvider:       byProvider,
	}
}

// --- Time Series Response ---

type TimeSeriesResponse struct {
	Days []DailyMetricDTO `json:"days"`
}

type DailyMetricDTO struct {
	Date     string `json:"date"`
	Inbound  int64  `json:"inbound"`
	Outbound int64  `json:"outbound"`
	Success  int64  `json:"success"`
	Failed   int64  `json:"failed"`
}

func ToTimeSeriesResponse(dailyMetrics []entity.DailyMetric) *TimeSeriesResponse {
	days := make([]DailyMetricDTO, 0, len(dailyMetrics))
	for _, d := range dailyMetrics {
		days = append(days, DailyMetricDTO{
			Date:     d.Date,
			Inbound:  d.Inbound,
			Outbound: d.Outbound,
			Success:  d.Success,
			Failed:   d.Failed,
		})
	}
	return &TimeSeriesResponse{Days: days}
}

// --- Environment Breakdown Response ---

type EnvironmentBreakdownResponse struct {
	Environments []EnvironmentMetricDTO `json:"environments"`
}

type EnvironmentMetricDTO struct {
	EnvironmentID uuid.UUID `json:"environment_id"`
	Total         int64     `json:"total"`
	Inbound       int64     `json:"inbound"`
	Outbound      int64     `json:"outbound"`
	Success       int64     `json:"success"`
	Failed        int64     `json:"failed"`
}

func ToEnvironmentBreakdownResponse(envMetrics []entity.EnvironmentMetric) *EnvironmentBreakdownResponse {
	envs := make([]EnvironmentMetricDTO, 0, len(envMetrics))
	for _, e := range envMetrics {
		envs = append(envs, EnvironmentMetricDTO{
			EnvironmentID: e.EnvironmentID,
			Total:         e.Total,
			Inbound:       e.Inbound,
			Outbound:      e.Outbound,
			Success:       e.Success,
			Failed:        e.Failed,
		})
	}
	return &EnvironmentBreakdownResponse{Environments: envs}
}
