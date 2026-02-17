package repository

import (
	"context"
	"time"

	"CONVERDA/internal/health/dto"

	"gorm.io/gorm"
)

// WebhookHealthAdapter satisfies the health.WebhookHealthReader interface using the webhook_logs table.
type WebhookHealthAdapter struct {
	db *gorm.DB
}

func NewWebhookHealthAdapter(db *gorm.DB) *WebhookHealthAdapter {
	return &WebhookHealthAdapter{db: db}
}

func (r *WebhookHealthAdapter) GetHealthStats(ctx context.Context, from, to time.Time) (*dto.WebhookHealth, error) {
	type result struct {
		TotalDispatched int
		SuccessCount    int
		FailedCount     int
		PendingRetries  int
	}

	var res result
	err := r.db.WithContext(ctx).Table("webhook_logs").
		Where("created_at BETWEEN ? AND ?", from, to).
		Select(`
			COUNT(id) AS total_dispatched,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) AS success_count,
			SUM(CASE WHEN status = 'failed' AND (retry_count >= max_retries OR next_retry_at IS NULL) THEN 1 ELSE 0 END) AS failed_count,
			SUM(CASE WHEN status = 'failed' AND retry_count < max_retries AND next_retry_at IS NOT NULL THEN 1 ELSE 0 END) AS pending_retries
		`).
		Scan(&res).Error
	if err != nil {
		return nil, err
	}

	var rate float64
	if res.TotalDispatched > 0 {
		rate = float64(res.SuccessCount) / float64(res.TotalDispatched) * 100
	}

	return &dto.WebhookHealth{
		TotalDispatched: res.TotalDispatched,
		SuccessCount:    res.SuccessCount,
		FailedCount:     res.FailedCount,
		PendingRetries:  res.PendingRetries,
		SuccessRate:     rate,
	}, nil
}
