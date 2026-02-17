package service

import (
	"context"

	"CONVERDA/internal/workflow/domain/model/entity"
)

// StepHandler executes a single step type within a workflow.
type StepHandler interface {
	// Execute runs the step logic. Returns nil on success, error on failure.
	Execute(ctx context.Context, step *entity.WorkflowStep, exec *entity.WorkflowExecution,
		stepExec *entity.StepExecution) error
}
