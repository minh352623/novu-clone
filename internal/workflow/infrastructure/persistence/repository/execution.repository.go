package repository

import (
	"context"
	"time"

	"CONVERDA/internal/workflow/domain/model/entity"
	domainRepo "CONVERDA/internal/workflow/domain/repository"
	"CONVERDA/internal/workflow/infrastructure/persistence/mapper"
	"CONVERDA/internal/workflow/infrastructure/persistence/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type executionRepository struct {
	db *gorm.DB
}

func NewExecutionRepository(db *gorm.DB) domainRepo.ExecutionRepository {
	return &executionRepository{db: db}
}

func (r *executionRepository) CreateExecution(ctx context.Context, exec *entity.WorkflowExecution) error {
	m := mapper.ToExecutionModel(exec)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *executionRepository) GetExecution(ctx context.Context, id uuid.UUID) (*entity.WorkflowExecution, error) {
	var m model.WorkflowExecutionModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return mapper.ToExecutionDomain(&m), nil
}

func (r *executionRepository) UpdateExecution(ctx context.Context, exec *entity.WorkflowExecution) error {
	m := mapper.ToExecutionModel(exec)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *executionRepository) CreateStepExecution(ctx context.Context, stepExec *entity.StepExecution) error {
	m := mapper.ToStepExecutionModel(stepExec)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *executionRepository) UpdateStepExecution(ctx context.Context, stepExec *entity.StepExecution) error {
	m := mapper.ToStepExecutionModel(stepExec)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *executionRepository) GetPendingScheduledSteps(ctx context.Context, limit int) ([]*entity.StepExecution, error) {
	var models []model.StepExecutionModel
	err := r.db.WithContext(ctx).
		Where("status = ? AND scheduled_at <= ?", entity.StepStatusScheduled, time.Now()).
		Order("scheduled_at ASC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	result := make([]*entity.StepExecution, 0, len(models))
	for i := range models {
		result = append(result, mapper.ToStepExecutionDomain(&models[i]))
	}
	return result, nil
}

func (r *executionRepository) GetStepExecutionsByExecution(ctx context.Context, executionID uuid.UUID) ([]*entity.StepExecution, error) {
	var models []model.StepExecutionModel
	err := r.db.WithContext(ctx).
		Where("execution_id = ?", executionID).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	result := make([]*entity.StepExecution, 0, len(models))
	for i := range models {
		result = append(result, mapper.ToStepExecutionDomain(&models[i]))
	}
	return result, nil
}

// --- Digest support ---

func (r *executionRepository) BufferDigestEvent(ctx context.Context, event *entity.DigestEvent) error {
	m := mapper.ToDigestEventModel(event)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *executionRepository) FlushDigestEvents(ctx context.Context, stepID, executionID uuid.UUID) ([]*entity.DigestEvent, error) {
	var models []model.DigestEventModel
	err := r.db.WithContext(ctx).
		Where("step_id = ? AND execution_id = ?", stepID, executionID).
		Order("created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	// Delete flushed events
	if len(models) > 0 {
		r.db.WithContext(ctx).
			Where("step_id = ? AND execution_id = ?", stepID, executionID).
			Delete(&model.DigestEventModel{})
	}

	result := make([]*entity.DigestEvent, 0, len(models))
	for i := range models {
		result = append(result, mapper.ToDigestEventDomain(&models[i]))
	}
	return result, nil
}

func (r *executionRepository) GetDigestingSteps(ctx context.Context, limit int) ([]*entity.StepExecution, error) {
	var models []model.StepExecutionModel
	err := r.db.WithContext(ctx).
		Where("status = ? AND scheduled_at <= ?", entity.StepStatusDigesting, time.Now()).
		Order("scheduled_at ASC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	result := make([]*entity.StepExecution, 0, len(models))
	for i := range models {
		result = append(result, mapper.ToStepExecutionDomain(&models[i]))
	}
	return result, nil
}
