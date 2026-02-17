package repository

import (
	"context"
	"fmt"
	"time"

	"CONVERDA/internal/messaging/domain/model/entity"
	domainRepo "CONVERDA/internal/messaging/domain/repository"
	"CONVERDA/internal/messaging/infrastructure/persistence/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type assignmentLogRepository struct {
	db *gorm.DB
}

func NewAssignmentLogRepository(db *gorm.DB) domainRepo.AssignmentLogRepository {
	return &assignmentLogRepository{db: db}
}

func (r *assignmentLogRepository) Create(ctx context.Context, log *entity.AssignmentLog) error {
	m := &model.AssignmentLogModel{
		ID:                 log.ID,
		ThreadID:           log.ThreadID,
		AssignedToMemberID: log.AssignedToMemberID,
		AssignedAt:         log.AssignedAt,
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *assignmentLogRepository) GetLastByThread(ctx context.Context, threadID uuid.UUID) (*entity.AssignmentLog, error) {
	var m model.AssignmentLogModel
	err := r.db.WithContext(ctx).
		Where("thread_id = ?", threadID).
		Order("assigned_at DESC").
		First(&m).Error
	if err != nil {
		return nil, err
	}
	return &entity.AssignmentLog{
		ID:                  m.ID,
		ThreadID:            m.ThreadID,
		AssignedToMemberID:  m.AssignedToMemberID,
		AssignedAt:          m.AssignedAt,
		ResolvedAt:          m.ResolvedAt,
		ResponseTimeSeconds: m.ResponseTimeSeconds,
	}, nil
}

func (r *assignmentLogRepository) GetByThread(ctx context.Context, threadID uuid.UUID) ([]*entity.AssignmentLog, error) {
	var models []model.AssignmentLogModel
	err := r.db.WithContext(ctx).
		Where("thread_id = ?", threadID).
		Order("assigned_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	result := make([]*entity.AssignmentLog, len(models))
	for i, m := range models {
		result[i] = &entity.AssignmentLog{
			ID:                  m.ID,
			ThreadID:            m.ThreadID,
			AssignedToMemberID:  m.AssignedToMemberID,
			AssignedAt:          m.AssignedAt,
			ResolvedAt:          m.ResolvedAt,
			ResponseTimeSeconds: m.ResponseTimeSeconds,
		}
	}
	return result, nil
}

func (r *assignmentLogRepository) Update(ctx context.Context, log *entity.AssignmentLog) error {
	m := &model.AssignmentLogModel{
		ID:                  log.ID,
		ResolvedAt:          log.ResolvedAt,
		ResponseTimeSeconds: log.ResponseTimeSeconds,
	}
	return r.db.WithContext(ctx).Model(m).Updates(m).Error
}

func (r *assignmentLogRepository) GetTeamStats(ctx context.Context, envID uuid.UUID, from, to time.Time, slaSeconds int) (*entity.TeamStats, error) {
	var stats entity.TeamStats
	var distribution entity.ResponseTimeDistribution

	// Total Conversations (Unique threads assigned in period)
	err := r.db.WithContext(ctx).Table("assignment_logs").
		Joins("JOIN threads ON threads.id = assignment_logs.thread_id").
		Where("threads.environment_id = ? AND assignment_logs.assigned_at BETWEEN ? AND ?", envID, from, to).
		Select("COUNT(DISTINCT assignment_logs.thread_id)").Scan(&stats.TotalConversations).Error
	if err != nil {
		return nil, err
	}

	// Resolved Count, Avg Response Time & Distribution
	type ResStats struct {
		Count      int
		Avg        float64
		Under5m    int
		From5To15m int
		Over15m    int
	}
	var res ResStats
	err = r.db.WithContext(ctx).Table("assignment_logs").
		Joins("JOIN threads ON threads.id = assignment_logs.thread_id").
		Where("threads.environment_id = ? AND assignment_logs.resolved_at BETWEEN ? AND ?", envID, from, to).
		Select(`
			COUNT(assignment_logs.id) as count, 
			COALESCE(AVG(assignment_logs.response_time_seconds), 0) as avg,
			COUNT(CASE WHEN assignment_logs.response_time_seconds < 300 THEN 1 END) as under_5m,
			COUNT(CASE WHEN assignment_logs.response_time_seconds BETWEEN 300 AND 900 THEN 1 END) as from_5_to_15m,
			COUNT(CASE WHEN assignment_logs.response_time_seconds > 900 THEN 1 END) as over_15m
		`).Scan(&res).Error

	if err != nil {
		return nil, err
	}
	stats.ResolvedCount = res.Count
	stats.AvgResponseTime = res.Avg
	distribution.Under5m = res.Under5m
	distribution.From5To15m = res.From5To15m
	distribution.Over15m = res.Over15m
	stats.ResponseTimeDistribution = distribution

	// SLA Compliance
	var compliantCount int64
	r.db.WithContext(ctx).Table("assignment_logs").
		Joins("JOIN threads ON threads.id = assignment_logs.thread_id").
		Where("threads.environment_id = ? AND assignment_logs.resolved_at BETWEEN ? AND ? AND assignment_logs.response_time_seconds <= ?", envID, from, to, slaSeconds).
		Count(&compliantCount)

	if stats.ResolvedCount > 0 {
		stats.SLAComplianceRate = float64(compliantCount) / float64(stats.ResolvedCount) * 100
	} else {
		stats.SLAComplianceRate = 100 // No resolved = 100% compliant? or 0? 100 is friendlier.
	}

	return &stats, nil
}

func (r *assignmentLogRepository) GetAgentStats(ctx context.Context, envID uuid.UUID, memberID uuid.UUID, from, to time.Time, slaSeconds int) (*entity.AgentStats, error) {
	stats := &entity.AgentStats{
		MemberID: memberID,
	}
	var distribution entity.ResponseTimeDistribution

	// Total Assigned
	var totalAssigned int64
	r.db.WithContext(ctx).Table("assignment_logs").
		Joins("JOIN threads ON threads.id = assignment_logs.thread_id").
		Where("threads.environment_id = ? AND assignment_logs.assigned_to_member_id = ? AND assignment_logs.assigned_at BETWEEN ? AND ?", envID, memberID, from, to).
		Count(&totalAssigned)
	stats.TotalAssigned = int(totalAssigned)

	// Total Resolved, Avg Response Time & Distribution
	type ResStats struct {
		Count      int
		Avg        float64
		Under5m    int
		From5To15m int
		Over15m    int
	}
	var res ResStats
	r.db.WithContext(ctx).Table("assignment_logs").
		Joins("JOIN threads ON threads.id = assignment_logs.thread_id").
		Where("threads.environment_id = ? AND assignment_logs.assigned_to_member_id = ? AND assignment_logs.resolved_at BETWEEN ? AND ?", envID, memberID, from, to).
		Select(`
			COUNT(assignment_logs.id) as count, 
			COALESCE(AVG(assignment_logs.response_time_seconds), 0) as avg,
			COUNT(CASE WHEN assignment_logs.response_time_seconds < 300 THEN 1 END) as under_5m,
			COUNT(CASE WHEN assignment_logs.response_time_seconds BETWEEN 300 AND 900 THEN 1 END) as from_5_to_15m,
			COUNT(CASE WHEN assignment_logs.response_time_seconds > 900 THEN 1 END) as over_15m
		`).Scan(&res)

	stats.TotalResolved = res.Count
	stats.AvgResponseTime = res.Avg
	distribution.Under5m = res.Under5m
	distribution.From5To15m = res.From5To15m
	distribution.Over15m = res.Over15m
	stats.ResponseTimeDistribution = distribution

	// Current Open Threads (Snapshot, ignore time range? Or threads assigned in range but not resolved?)
	// Usually "Current" means NOW.
	var openCount int64
	// Assigned but not Resolved (resolved_at IS NULL).
	// Note: If re-assigned, previous log remains unresolved? This needs stricter logic.
	// For now, simpler approach: Count where AssignedTo is Member AND Thread Status is 'assigned'.
	r.db.WithContext(ctx).Table("threads").
		Joins("JOIN thread_participants tp ON tp.thread_id = threads.id").
		Where("threads.environment_id = ? AND threads.status = 'assigned' AND tp.entity_type = 'user' AND tp.entity_id = ?", envID, memberID).
		Count(&openCount)
	stats.CurrentOpenThreads = int(openCount)
	stats.Workload = int(openCount)

	return stats, nil
}

func (r *assignmentLogRepository) GetActivityTimeline(ctx context.Context, envID uuid.UUID, memberID uuid.UUID, from, to time.Time, interval string) ([]*entity.ActivityPoint, error) {
	var points []*entity.ActivityPoint

	// interval: 'hour' or 'day'
	trunc := "hour"
	if interval == "day" {
		trunc = "day"
	}

	err := r.db.WithContext(ctx).Table("assignment_logs").
		Joins("JOIN threads ON threads.id = assignment_logs.thread_id").
		Where("threads.environment_id = ? AND assignment_logs.assigned_to_member_id = ? AND assignment_logs.resolved_at BETWEEN ? AND ?", envID, memberID, from, to).
		Select(fmt.Sprintf("date_trunc('%s', assignment_logs.resolved_at) as timestamp, count(assignment_logs.id) as value", trunc)).
		Group("timestamp").
		Order("timestamp ASC").
		Scan(&points).Error

	return points, err
}
