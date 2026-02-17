package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ThreadHealthAdapter satisfies the health.QueueReader interface using the threads table.
type ThreadHealthAdapter struct {
	db *gorm.DB
}

func NewThreadHealthAdapter(db *gorm.DB) *ThreadHealthAdapter {
	return &ThreadHealthAdapter{db: db}
}

func (r *ThreadHealthAdapter) CountByStatus(ctx context.Context, envID uuid.UUID, status string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("threads").
		Where("environment_id = ? AND status = ?", envID, status).
		Count(&count).Error
	return count, err
}

func (r *ThreadHealthAdapter) CountOverdue(ctx context.Context, envID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("threads").
		Where("environment_id = ? AND is_overdue = ? AND status != ?", envID, true, "resolved").
		Count(&count).Error
	return count, err
}

func (r *ThreadHealthAdapter) AvgWaitTime(ctx context.Context, envID uuid.UUID) (float64, error) {
	var avg *float64
	err := r.db.WithContext(ctx).Table("threads").
		Where("environment_id = ? AND status = ?", envID, "unassigned").
		Select("COALESCE(AVG(EXTRACT(EPOCH FROM (NOW() - created_at))), 0)").
		Scan(&avg).Error
	if err != nil {
		return 0, err
	}
	if avg == nil {
		return 0, nil
	}
	return *avg, nil
}
