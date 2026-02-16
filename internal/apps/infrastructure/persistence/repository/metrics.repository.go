package repository

import (
	"context"
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
