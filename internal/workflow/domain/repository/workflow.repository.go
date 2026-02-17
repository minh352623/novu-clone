package repository

import (
	"context"

	"CONVERDA/internal/workflow/domain/model/entity"

	"github.com/google/uuid"
)

// WorkflowRepository manages Workflow and WorkflowStep persistence.
type WorkflowRepository interface {
	Create(ctx context.Context, workflow *entity.Workflow) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Workflow, error)
	ListByEnvironment(ctx context.Context, envID uuid.UUID) ([]*entity.Workflow, error)
	Update(ctx context.Context, workflow *entity.Workflow) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByTrigger(ctx context.Context, envID uuid.UUID, triggerIdentifier string) (*entity.Workflow, error)

	// Step operations
	AddStep(ctx context.Context, step *entity.WorkflowStep) error
	UpdateStep(ctx context.Context, step *entity.WorkflowStep) error
	DeleteStep(ctx context.Context, stepID uuid.UUID) error
	GetStepsByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]entity.WorkflowStep, error)
}
