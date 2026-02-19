package impl

import (
	"context"
	"fmt"
	"time"

	"CONVERDA/internal/workflow/controller/dto"
	"CONVERDA/internal/workflow/domain"
	"CONVERDA/internal/workflow/domain/model/entity"
	"CONVERDA/internal/workflow/domain/repository"

	"github.com/google/uuid"
)

type workflowServiceImpl struct {
	workflowRepo repository.WorkflowRepository
	uow          repository.WorkflowUnitOfWork
}

func NewWorkflowService(workflowRepo repository.WorkflowRepository, uow repository.WorkflowUnitOfWork) *workflowServiceImpl {
	return &workflowServiceImpl{
		workflowRepo: workflowRepo,
		uow:          uow,
	}
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

	if err := s.uow.Execute(ctx, func(tx repository.WorkflowTxRepository) error {
		return tx.Workflows().Create(ctx, wf)
	}); err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	return dto.ToWorkflowResponse(wf), nil
}

func (s *workflowServiceImpl) GetWorkflow(ctx context.Context, id uuid.UUID) (*dto.WorkflowResponse, error) {
	wf, err := s.workflowRepo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrWorkflowNotFound
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
	var wf *entity.Workflow
	if err := s.uow.Execute(ctx, func(tx repository.WorkflowTxRepository) error {
		var err error
		wf, err = tx.Workflows().GetByID(ctx, id)
		if err != nil {
			return domain.ErrWorkflowNotFound
		}

		if req.Name != nil {
			wf.Name = *req.Name
		}
		if req.TriggerIdentifier != nil {
			wf.TriggerIdentifier = *req.TriggerIdentifier
		}
		wf.UpdatedAt = time.Now()

		return tx.Workflows().Update(ctx, wf)
	}); err != nil {
		return nil, err
	}

	return dto.ToWorkflowResponse(wf), nil
}

func (s *workflowServiceImpl) DeleteWorkflow(ctx context.Context, id uuid.UUID) error {
	return s.uow.Execute(ctx, func(tx repository.WorkflowTxRepository) error {
		wf, err := tx.Workflows().GetByID(ctx, id)
		if err != nil {
			return domain.ErrWorkflowNotFound
		}
		if wf.IsActive {
			return domain.ErrWorkflowActive
		}
		return tx.Workflows().Delete(ctx, id)
	})
}

func (s *workflowServiceImpl) ToggleWorkflow(ctx context.Context, id uuid.UUID) (*dto.WorkflowResponse, error) {
	if err := s.uow.Execute(ctx, func(tx repository.WorkflowTxRepository) error {
		wf, err := tx.Workflows().GetByID(ctx, id)
		if err != nil {
			return domain.ErrWorkflowNotFound
		}

		wf.IsActive = !wf.IsActive
		wf.UpdatedAt = time.Now()

		return tx.Workflows().Update(ctx, wf)
	}); err != nil {
		return nil, err
	}

	wf, _ := s.workflowRepo.GetByID(ctx, id)
	return dto.ToWorkflowResponse(wf), nil
}

// --- Step operations ---

func (s *workflowServiceImpl) AddStep(ctx context.Context, workflowID uuid.UUID, req dto.CreateStepRequest) (*dto.StepResponse, error) {
	var step *entity.WorkflowStep
	if err := s.uow.Execute(ctx, func(tx repository.WorkflowTxRepository) error {
		// Verify workflow exists
		if _, err := tx.Workflows().GetByID(ctx, workflowID); err != nil {
			return domain.ErrWorkflowNotFound
		}

		step = &entity.WorkflowStep{
			ID:           uuid.New(),
			WorkflowID:   workflowID,
			ParentStepID: req.ParentStepID,
			StepType:     req.StepType,
			Config:       req.Config,
			Order:        req.Order,
			CreatedAt:    time.Now(),
		}

		return tx.Workflows().AddStep(ctx, step)
	}); err != nil {
		return nil, err
	}

	resp := dto.ToStepResponse(step)
	return &resp, nil
}

func (s *workflowServiceImpl) UpdateStep(ctx context.Context, stepID uuid.UUID, req dto.UpdateStepRequest) (*dto.StepResponse, error) {
	step := &entity.WorkflowStep{
		ID: stepID,
	}

	if err := s.uow.Execute(ctx, func(tx repository.WorkflowTxRepository) error {
		if req.StepType != nil {
			step.StepType = *req.StepType
		}
		if req.Config != nil {
			step.Config = req.Config
		}
		if req.Order != nil {
			step.Order = *req.Order
		}

		return tx.Workflows().UpdateStep(ctx, step)
	}); err != nil {
		return nil, fmt.Errorf("failed to update step %s: %w", stepID, err)
	}

	resp := dto.ToStepResponse(step)
	return &resp, nil
}

func (s *workflowServiceImpl) DeleteStep(ctx context.Context, stepID uuid.UUID) error {
	return s.uow.Execute(ctx, func(tx repository.WorkflowTxRepository) error {
		return tx.Workflows().DeleteStep(ctx, stepID)
	})
}
