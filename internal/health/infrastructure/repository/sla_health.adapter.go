package repository

import (
	"context"
	"time"

	"CONVERDA/internal/health/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SLAHealthAdapter satisfies the health.SLAReader interface using the assignment_logs table.
type SLAHealthAdapter struct {
	db *gorm.DB
}

func NewSLAHealthAdapter(db *gorm.DB) *SLAHealthAdapter {
	return &SLAHealthAdapter{db: db}
}

func (r *SLAHealthAdapter) GetSLACompliance(ctx context.Context, envID uuid.UUID, slaSeconds int, from, to time.Time) (*dto.SLACompliance, error) {
	type result struct {
		TotalResolved int
		WithinSLA     int
	}

	var res result
	err := r.db.WithContext(ctx).Table("assignment_logs").
		Joins("JOIN threads ON threads.id = assignment_logs.thread_id").
		Where("threads.environment_id = ? AND assignment_logs.resolved_at BETWEEN ? AND ?", envID, from, to).
		Select(`
			COUNT(assignment_logs.id) AS total_resolved,
			SUM(CASE WHEN assignment_logs.response_time_seconds <= ? THEN 1 ELSE 0 END) AS within_sla
		`, slaSeconds).
		Scan(&res).Error
	if err != nil {
		return nil, err
	}

	breached := res.TotalResolved - res.WithinSLA
	var rate float64
	if res.TotalResolved > 0 {
		rate = float64(res.WithinSLA) / float64(res.TotalResolved) * 100
	}

	return &dto.SLACompliance{
		TotalResolved:  res.TotalResolved,
		WithinSLA:      res.WithinSLA,
		Breached:       breached,
		ComplianceRate: rate,
	}, nil
}
