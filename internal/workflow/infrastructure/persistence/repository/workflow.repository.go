package repository

import (
	"context"

	"CONVERDA/internal/workflow/domain/model/entity"
	domainRepo "CONVERDA/internal/workflow/domain/repository"
	"CONVERDA/internal/workflow/infrastructure/persistence/mapper"
	"CONVERDA/internal/workflow/infrastructure/persistence/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type workflowRepository struct {
	db *gorm.DB
}

func NewWorkflowRepository(db *gorm.DB) domainRepo.WorkflowRepository {
	return &workflowRepository{db: db}
}

func (r *workflowRepository) Create(ctx context.Context, workflow *entity.Workflow) error {
	m := mapper.ToWorkflowModel(workflow)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *workflowRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Workflow, error) {
	var m model.WorkflowModel
	err := r.db.WithContext(ctx).Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order(`"order" ASC`)
	}).First(&m, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return mapper.ToWorkflowDomain(&m), nil
}

func (r *workflowRepository) ListByEnvironment(ctx context.Context, envID uuid.UUID) ([]*entity.Workflow, error) {
	var models []model.WorkflowModel
	err := r.db.WithContext(ctx).
		Where("environment_id = ?", envID).
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	result := make([]*entity.Workflow, 0, len(models))
	for i := range models {
		result = append(result, mapper.ToWorkflowDomain(&models[i]))
	}
	return result, nil
}

func (r *workflowRepository) Update(ctx context.Context, workflow *entity.Workflow) error {
	m := mapper.ToWorkflowModel(workflow)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *workflowRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.WorkflowModel{}, "id = ?", id).Error
}

func (r *workflowRepository) GetByTrigger(ctx context.Context, envID uuid.UUID, triggerIdentifier string) (*entity.Workflow, error) {
	var m model.WorkflowModel
	err := r.db.WithContext(ctx).
		Where("environment_id = ? AND trigger_identifier = ? AND is_active = ?", envID, triggerIdentifier, true).
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		First(&m).Error
	if err != nil {
		return nil, err
	}
	return mapper.ToWorkflowDomain(&m), nil
}

// --- Step operations ---

func (r *workflowRepository) AddStep(ctx context.Context, step *entity.WorkflowStep) error {
	m := mapper.ToStepModel(step)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *workflowRepository) UpdateStep(ctx context.Context, step *entity.WorkflowStep) error {
	m := mapper.ToStepModel(step)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *workflowRepository) DeleteStep(ctx context.Context, stepID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.WorkflowStepModel{}, "id = ?", stepID).Error
}

func (r *workflowRepository) GetStepsByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]entity.WorkflowStep, error) {
	var models []model.WorkflowStepModel
	err := r.db.WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Order(`"order" ASC`).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	result := make([]entity.WorkflowStep, 0, len(models))
	for i := range models {
		result = append(result, *mapper.ToStepDomain(&models[i]))
	}
	return result, nil
}
