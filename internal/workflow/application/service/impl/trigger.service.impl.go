package impl

import (
	"context"
	"fmt"
	"sort"
	"time"

	"CONVERDA/global"
	"CONVERDA/internal/workflow/application/service"
	"CONVERDA/internal/workflow/domain"
	"CONVERDA/internal/workflow/domain/model/entity"
	"CONVERDA/internal/workflow/domain/repository"

	"github.com/google/uuid"
)

type triggerServiceImpl struct {
	workflowRepo   repository.WorkflowRepository
	execRepo       repository.ExecutionRepository
	uow            repository.WorkflowUnitOfWork
	channelHandler service.StepHandler
	delayHandler   service.StepHandler
}

func NewTriggerService(
	workflowRepo repository.WorkflowRepository,
	execRepo repository.ExecutionRepository,
	uow repository.WorkflowUnitOfWork,
	channelHandler service.StepHandler,
	delayHandler service.StepHandler,
) service.TriggerService {
	return &triggerServiceImpl{
		workflowRepo:   workflowRepo,
		execRepo:       execRepo,
		uow:            uow,
		channelHandler: channelHandler,
		delayHandler:   delayHandler,
	}
}

func (s *triggerServiceImpl) Trigger(ctx context.Context, envID uuid.UUID, triggerIdentifier string,
	subscriberKey string, payload map[string]interface{}) error {

	// 1. Find active workflow
	workflow, err := s.workflowRepo.GetByTrigger(ctx, envID, triggerIdentifier)
	if err != nil {
		return domain.ErrWorkflowNotFound
	}

	if len(workflow.Steps) == 0 {
		return fmt.Errorf("workflow %s has no steps", workflow.ID)
	}

	// Sort steps by order
	steps := make([]entity.WorkflowStep, len(workflow.Steps))
	copy(steps, workflow.Steps)
	sort.Slice(steps, func(i, j int) bool {
		return steps[i].Order < steps[j].Order
	})

	return s.uow.Execute(ctx, func(txRepo repository.WorkflowTxRepository) error {
		txExecRepo := txRepo.Executions()

		// 2. Create WorkflowExecution
		now := time.Now()
		firstStepID := steps[0].ID
		exec := &entity.WorkflowExecution{
			ID:             uuid.New(),
			WorkflowID:     workflow.ID,
			SubscriberKey:  subscriberKey,
			TriggerPayload: payload,
			Status:         entity.WorkflowExecutionRunning,
			CurrentStepID:  &firstStepID,
			StartedAt:      now,
		}

		if err := txExecRepo.CreateExecution(ctx, exec); err != nil {
			return fmt.Errorf("failed to create execution: %w", err)
		}

		// 3. Create StepExecution rows for all steps
		stepExecs := make([]*entity.StepExecution, 0, len(steps))
		for _, step := range steps {
			se := &entity.StepExecution{
				ID:          uuid.New(),
				ExecutionID: exec.ID,
				StepID:      step.ID,
				Status:      entity.StepStatusPending,
			}
			if err := txExecRepo.CreateStepExecution(ctx, se); err != nil {
				return fmt.Errorf("failed to create step execution: %w", err)
			}
			stepExecs = append(stepExecs, se)
		}

		// 4. Execute steps starting from the first
		return s.executeStepsFrom(ctx, txExecRepo, exec, steps, stepExecs, 0)
	})
}

// executeStepsFrom processes steps sequentially starting from the given index.
// Stops when a Delay step is encountered (it will be picked up by the worker) or on error.
func (s *triggerServiceImpl) executeStepsFrom(ctx context.Context, execRepo repository.ExecutionRepository,
	exec *entity.WorkflowExecution, steps []entity.WorkflowStep,
	stepExecs []*entity.StepExecution, fromIndex int) error {

	for i := fromIndex; i < len(steps); i++ {
		step := &steps[i]
		stepExec := stepExecs[i]

		// Update current step pointer
		exec.CurrentStepID = &step.ID
		_ = execRepo.UpdateExecution(ctx, exec)

		// Mark step as running
		now := time.Now()
		stepExec.Status = entity.StepStatusRunning
		stepExec.StartedAt = &now
		_ = execRepo.UpdateStepExecution(ctx, stepExec)

		var handler service.StepHandler
		switch step.StepType {
		case entity.StepTypeChannel:
			handler = s.channelHandler
		case entity.StepTypeDelay:
			handler = s.delayHandler
		default:
			global.Logger.Error("workflow_trigger: unknown step type", "stepType", step.StepType, "executionID", exec.ID.String())
			return s.markStepFailed(ctx, execRepo, stepExec, exec, "unknown step type: "+step.StepType)
		}

		if err := handler.Execute(ctx, step, exec, stepExec); err != nil {
			global.Logger.Error("workflow_trigger: step execution failed", "stepID", step.ID.String(), "executionID", exec.ID.String(), "error", err)
			return s.markStepFailed(ctx, execRepo, stepExec, exec, err.Error())
		}

		// Delay steps set their own status to "scheduled" — stop processing
		if step.StepType == entity.StepTypeDelay {
			global.Logger.Info("workflow_trigger: delay step scheduled, pausing execution", "executionID", exec.ID.String())
			return nil
		}

		// Channel step success — mark completed and continue
		completedAt := time.Now()
		stepExec.Status = entity.StepStatusCompleted
		stepExec.CompletedAt = &completedAt
		if err := execRepo.UpdateStepExecution(ctx, stepExec); err != nil {
			global.Logger.Warn("workflow_trigger: failed to update step execution status", "stepExecID", stepExec.ID.String(), "error", err)
		}
	}

	// All steps completed
	completedAt := time.Now()
	exec.Status = entity.WorkflowExecutionCompleted
	exec.CompletedAt = &completedAt
	if err := execRepo.UpdateExecution(ctx, exec); err != nil {
		global.Logger.Warn("workflow_trigger: failed to update execution status", "executionID", exec.ID.String(), "error", err)
		return fmt.Errorf("failed to complete execution: %w", err)
	}

	global.Logger.Info("workflow_trigger: execution completed", "executionID", exec.ID.String())
	return nil
}

func (s *triggerServiceImpl) markStepFailed(ctx context.Context, execRepo repository.ExecutionRepository,
	stepExec *entity.StepExecution, exec *entity.WorkflowExecution, reason string) error {

	now := time.Now()
	stepExec.Status = entity.StepStatusFailed
	stepExec.CompletedAt = &now
	stepExec.Output = map[string]interface{}{"error": reason}
	if err := execRepo.UpdateStepExecution(ctx, stepExec); err != nil {
		global.Logger.Warn("workflow_trigger: failed to mark step as failed", "stepExecID", stepExec.ID.String(), "error", err)
	}

	exec.Status = entity.WorkflowExecutionFailed
	exec.CompletedAt = &now
	if err := execRepo.UpdateExecution(ctx, exec); err != nil {
		global.Logger.Warn("workflow_trigger: failed to mark execution as failed", "executionID", exec.ID.String(), "error", err)
	}

	return fmt.Errorf("step failed: %s", reason)
}
