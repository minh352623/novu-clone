package impl

import (
	"context"
	"fmt"
	"time"

	"CONVERDA/internal/workflow/controller/dto"
	"CONVERDA/internal/workflow/domain/model/entity"
	"CONVERDA/internal/workflow/domain/repository"

	"github.com/google/uuid"
)

type workflowServiceImpl struct {
	workflowRepo repository.WorkflowRepository
}

func NewWorkflowService(workflowRepo repository.WorkflowRepository) *workflowServiceImpl {
	return &workflowServiceImpl{workflowRepo: workflowRepo}
}

func (s *workflowServiceImpl) CreateWorkflow(ctx context.Context, envID uuid.UUID, req dto.CreateWorkflowRequest) (*dto.WorkflowResponse, error) {
	wf := &entity.Workflow{
		ID:                uuid.New(),
		EnvironmentID:     envID,
		Name:              req.Name,
		TriggerIdentifier: req.TriggerIdentifier,
		IsActive:          false,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.workflowRepo.Create(ctx, wf); err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	return dto.ToWorkflowResponse(wf), nil
}

func (s *workflowServiceImpl) GetWorkflow(ctx context.Context, id uuid.UUID) (*dto.WorkflowResponse, error) {
	wf, err := s.workflowRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow %s: %w", id, err)
	}
	return dto.ToWorkflowResponse(wf), nil
}

func (s *workflowServiceImpl) ListWorkflows(ctx context.Context, envID uuid.UUID) ([]*dto.WorkflowResponse, error) {
	workflows, err := s.workflowRepo.ListByEnvironment(ctx, envID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows for env %s: %w", envID, err)
	}
	return dto.ToWorkflowResponseList(workflows), nil
}

func (s *workflowServiceImpl) UpdateWorkflow(ctx context.Context, id uuid.UUID, req dto.UpdateWorkflowRequest) (*dto.WorkflowResponse, error) {
	wf, err := s.workflowRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow %s for update: %w", id, err)
	}

	if req.Name != nil {
		wf.Name = *req.Name
	}
	if req.TriggerIdentifier != nil {
		wf.TriggerIdentifier = *req.TriggerIdentifier
	}
	wf.UpdatedAt = time.Now()

	if err := s.workflowRepo.Update(ctx, wf); err != nil {
		return nil, fmt.Errorf("failed to update workflow %s: %w", id, err)
	}

	return dto.ToWorkflowResponse(wf), nil
}

func (s *workflowServiceImpl) DeleteWorkflow(ctx context.Context, id uuid.UUID) error {
	if err := s.workflowRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete workflow %s: %w", id, err)
	}
	return nil
}

func (s *workflowServiceImpl) ToggleWorkflow(ctx context.Context, id uuid.UUID) (*dto.WorkflowResponse, error) {
	wf, err := s.workflowRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow %s for toggle: %w", id, err)
	}

	wf.IsActive = !wf.IsActive
	wf.UpdatedAt = time.Now()

	if err := s.workflowRepo.Update(ctx, wf); err != nil {
		return nil, fmt.Errorf("failed to update workflow %s after toggle: %w", id, err)
	}

	return dto.ToWorkflowResponse(wf), nil
}

// --- Step operations ---

func (s *workflowServiceImpl) AddStep(ctx context.Context, workflowID uuid.UUID, req dto.CreateStepRequest) (*dto.StepResponse, error) {
	// Verify workflow exists
	if _, err := s.workflowRepo.GetByID(ctx, workflowID); err != nil {
		return nil, fmt.Errorf("workflow %s not found: %w", workflowID, err)
	}

	step := &entity.WorkflowStep{
		ID:           uuid.New(),
		WorkflowID:   workflowID,
		ParentStepID: req.ParentStepID,
		StepType:     req.StepType,
		Config:       req.Config,
		Order:        req.Order,
		CreatedAt:    time.Now(),
	}

	if err := s.workflowRepo.AddStep(ctx, step); err != nil {
		return nil, fmt.Errorf("failed to add step to workflow %s: %w", workflowID, err)
	}

	resp := dto.ToStepResponse(step)
	return &resp, nil
}

func (s *workflowServiceImpl) UpdateStep(ctx context.Context, stepID uuid.UUID, req dto.UpdateStepRequest) (*dto.StepResponse, error) {
	// We need to find the step — get all steps from its workflow
	// For simplicity, do a direct update via the step fields
	step := &entity.WorkflowStep{
		ID: stepID,
	}

	if req.StepType != nil {
		step.StepType = *req.StepType
	}
	if req.Config != nil {
		step.Config = req.Config
	}
	if req.Order != nil {
		step.Order = *req.Order
	}

	if err := s.workflowRepo.UpdateStep(ctx, step); err != nil {
		return nil, fmt.Errorf("failed to update step %s: %w", stepID, err)
	}

	resp := dto.ToStepResponse(step)
	return &resp, nil
}

func (s *workflowServiceImpl) DeleteStep(ctx context.Context, stepID uuid.UUID) error {
	if err := s.workflowRepo.DeleteStep(ctx, stepID); err != nil {
		return fmt.Errorf("failed to delete step %s: %w", stepID, err)
	}
	return nil
}
