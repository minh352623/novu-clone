package repository

import (
	"context"
	"fmt"
	"time"

	"CONVERDA/internal/apps/domain/model/entity"
	"CONVERDA/internal/apps/domain/repository"
	"CONVERDA/internal/apps/infrastructure/persistence/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type usageMetricRepository struct {
	db *gorm.DB
}

func NewUsageMetricRepository(db *gorm.DB) repository.UsageMetricRepository {
	return &usageMetricRepository{db: db}
}

func (r *usageMetricRepository) Create(ctx context.Context, metric *entity.UsageMetric) error {
	m := &model.UsageMetricModel{
		ID:            metric.ID,
		AppID:         metric.AppID,
		EnvironmentID: metric.EnvironmentID,
		ProviderType:  metric.ProviderType,
		Direction:     metric.Direction,
		StatusCode:    metric.StatusCode,
		Timestamp:     metric.Timestamp,
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *usageMetricRepository) GetSummaryByApp(ctx context.Context, appID uuid.UUID, from, to time.Time) (map[string]int64, error) {
	var results []struct {
		ProviderType string
		Total        int64
	}

	err := r.db.WithContext(ctx).Model(&model.UsageMetricModel{}).
		Select("provider_type, count(*) as total").
		Where("app_id = ? AND timestamp BETWEEN ? AND ?", appID, from, to).
		Group("provider_type").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	summary := make(map[string]int64)
	for _, res := range results {
		summary[res.ProviderType] = res.Total
	}
	return summary, nil
}

func (r *usageMetricRepository) GetEnvironmentUsage(ctx context.Context, envID uuid.UUID, from, to time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.UsageMetricModel{}).
		Where("environment_id = ? AND timestamp BETWEEN ? AND ?", envID, from, to).
		Count(&count).Error
	return count, err
}

// --- US-RA-02.5: Detailed metrics ---

func (r *usageMetricRepository) GetDetailedSummary(ctx context.Context, appID uuid.UUID, from, to time.Time) (*entity.DetailedMetricsSummary, error) {
	// 1. Overall counts
	var overallResult struct {
		Total    int64
		Inbound  int64
		Outbound int64
		Success  int64
		Failed   int64
	}

	err := r.db.WithContext(ctx).Model(&model.UsageMetricModel{}).
		Select(fmt.Sprintf(`
			COUNT(*) as total,
			SUM(CASE WHEN direction = 'inbound' THEN 1 ELSE 0 END) as inbound,
			SUM(CASE WHEN direction = 'outbound' THEN 1 ELSE 0 END) as outbound,
			SUM(CASE WHEN status_code BETWEEN %d AND %d THEN 1 ELSE 0 END) as success,
			SUM(CASE WHEN status_code < %d OR status_code > %d THEN 1 ELSE 0 END) as failed
		`, entity.SuccessStatusMin, entity.SuccessStatusMax, entity.SuccessStatusMin, entity.SuccessStatusMax)).
		Where("app_id = ? AND timestamp BETWEEN ? AND ?", appID, from, to).
		Scan(&overallResult).Error

	if err != nil {
		return nil, err
	}

	// 2. Per-provider breakdown
	var providerResults []struct {
		ProviderType string
		Total        int64
		Success      int64
		Failed       int64
	}

	err = r.db.WithContext(ctx).Model(&model.UsageMetricModel{}).
		Select(fmt.Sprintf(`
			provider_type,
			COUNT(*) as total,
			SUM(CASE WHEN status_code BETWEEN %d AND %d THEN 1 ELSE 0 END) as success,
			SUM(CASE WHEN status_code < %d OR status_code > %d THEN 1 ELSE 0 END) as failed
		`, entity.SuccessStatusMin, entity.SuccessStatusMax, entity.SuccessStatusMin, entity.SuccessStatusMax)).
		Where("app_id = ? AND timestamp BETWEEN ? AND ?", appID, from, to).
		Group("provider_type").
		Scan(&providerResults).Error

	if err != nil {
		return nil, err
	}

	byProvider := make(map[string]entity.ProviderMetric)
	for _, p := range providerResults {
		byProvider[p.ProviderType] = entity.ProviderMetric{
			Total:   p.Total,
			Success: p.Success,
			Failed:  p.Failed,
		}
	}

	return &entity.DetailedMetricsSummary{
		TotalMessages:    overallResult.Total,
		InboundMessages:  overallResult.Inbound,
		OutboundMessages: overallResult.Outbound,
		SuccessCount:     overallResult.Success,
		FailureCount:     overallResult.Failed,
		ByProvider:       byProvider,
	}, nil
}

func (r *usageMetricRepository) GetDailyTimeSeries(ctx context.Context, appID uuid.UUID, envID *uuid.UUID, from, to time.Time) ([]entity.DailyMetric, error) {
	var results []struct {
		Date     string
		Inbound  int64
		Outbound int64
		Success  int64
		Failed   int64
	}

	query := r.db.WithContext(ctx).Model(&model.UsageMetricModel{}).
		Select(fmt.Sprintf(`
			TO_CHAR(timestamp, 'YYYY-MM-DD') as date,
			SUM(CASE WHEN direction = 'inbound' THEN 1 ELSE 0 END) as inbound,
			SUM(CASE WHEN direction = 'outbound' THEN 1 ELSE 0 END) as outbound,
			SUM(CASE WHEN status_code BETWEEN %d AND %d THEN 1 ELSE 0 END) as success,
			SUM(CASE WHEN status_code < %d OR status_code > %d THEN 1 ELSE 0 END) as failed
		`, entity.SuccessStatusMin, entity.SuccessStatusMax, entity.SuccessStatusMin, entity.SuccessStatusMax)).
		Where("app_id = ? AND timestamp BETWEEN ? AND ?", appID, from, to)

	if envID != nil {
		query = query.Where("environment_id = ?", *envID)
	}

	err := query.
		Group("TO_CHAR(timestamp, 'YYYY-MM-DD')").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	dailyMetrics := make([]entity.DailyMetric, 0, len(results))
	for _, r := range results {
		dailyMetrics = append(dailyMetrics, entity.DailyMetric{
			Date:     r.Date,
			Inbound:  r.Inbound,
			Outbound: r.Outbound,
			Success:  r.Success,
			Failed:   r.Failed,
		})
	}
	return dailyMetrics, nil
}

func (r *usageMetricRepository) GetEnvironmentBreakdown(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]entity.EnvironmentMetric, error) {
	var results []struct {
		EnvironmentID uuid.UUID
		Total         int64
		Inbound       int64
		Outbound      int64
		Success       int64
		Failed        int64
	}

	err := r.db.WithContext(ctx).Model(&model.UsageMetricModel{}).
		Select(fmt.Sprintf(`
			environment_id,
			COUNT(*) as total,
			SUM(CASE WHEN direction = 'inbound' THEN 1 ELSE 0 END) as inbound,
			SUM(CASE WHEN direction = 'outbound' THEN 1 ELSE 0 END) as outbound,
			SUM(CASE WHEN status_code BETWEEN %d AND %d THEN 1 ELSE 0 END) as success,
			SUM(CASE WHEN status_code < %d OR status_code > %d THEN 1 ELSE 0 END) as failed
		`, entity.SuccessStatusMin, entity.SuccessStatusMax, entity.SuccessStatusMin, entity.SuccessStatusMax)).
		Where("app_id = ? AND timestamp BETWEEN ? AND ?", appID, from, to).
		Group("environment_id").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	envMetrics := make([]entity.EnvironmentMetric, 0, len(results))
	for _, r := range results {
		envMetrics = append(envMetrics, entity.EnvironmentMetric{
			EnvironmentID: r.EnvironmentID,
			Total:         r.Total,
			Inbound:       r.Inbound,
			Outbound:      r.Outbound,
			Success:       r.Success,
			Failed:        r.Failed,
		})
	}
	return envMetrics, nil
}
