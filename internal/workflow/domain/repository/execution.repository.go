package repository

import (
	"context"

	"CONVERDA/internal/workflow/domain/model/entity"

	"github.com/google/uuid"
)

// ExecutionRepository manages WorkflowExecution and StepExecution persistence.
type ExecutionRepository interface {
	CreateExecution(ctx context.Context, exec *entity.WorkflowExecution) error
	GetExecution(ctx context.Context, id uuid.UUID) (*entity.WorkflowExecution, error)
	UpdateExecution(ctx context.Context, exec *entity.WorkflowExecution) error

	CreateStepExecution(ctx context.Context, stepExec *entity.StepExecution) error
	UpdateStepExecution(ctx context.Context, stepExec *entity.StepExecution) error
	GetPendingScheduledSteps(ctx context.Context, limit int) ([]*entity.StepExecution, error)
	GetStepExecutionsByExecution(ctx context.Context, executionID uuid.UUID) ([]*entity.StepExecution, error)

	// Digest support
	BufferDigestEvent(ctx context.Context, event *entity.DigestEvent) error
	FlushDigestEvents(ctx context.Context, stepID, executionID uuid.UUID) ([]*entity.DigestEvent, error)
	GetDigestingSteps(ctx context.Context, limit int) ([]*entity.StepExecution, error)
}
