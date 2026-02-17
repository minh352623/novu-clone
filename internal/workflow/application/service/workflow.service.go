package service

import (
	"context"

	"CONVERDA/internal/workflow/controller/dto"

	"github.com/google/uuid"
)

// WorkflowService handles workflow CRUD operations.
type WorkflowService interface {
	CreateWorkflow(ctx context.Context, envID uuid.UUID, req dto.CreateWorkflowRequest) (*dto.WorkflowResponse, error)
	GetWorkflow(ctx context.Context, id uuid.UUID) (*dto.WorkflowResponse, error)
	ListWorkflows(ctx context.Context, envID uuid.UUID) ([]*dto.WorkflowResponse, error)
	UpdateWorkflow(ctx context.Context, id uuid.UUID, req dto.UpdateWorkflowRequest) (*dto.WorkflowResponse, error)
	DeleteWorkflow(ctx context.Context, id uuid.UUID) error
	ToggleWorkflow(ctx context.Context, id uuid.UUID) (*dto.WorkflowResponse, error)

	// Step operations
	AddStep(ctx context.Context, workflowID uuid.UUID, req dto.CreateStepRequest) (*dto.StepResponse, error)
	UpdateStep(ctx context.Context, stepID uuid.UUID, req dto.UpdateStepRequest) (*dto.StepResponse, error)
	DeleteStep(ctx context.Context, stepID uuid.UUID) error
}
