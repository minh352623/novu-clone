package impl

import (
	"context"
	"fmt"
	"sort"
	"time"

	"CONVERDA/global"
	"CONVERDA/internal/workflow/application/service"
	"CONVERDA/internal/workflow/domain/model/entity"
	"CONVERDA/internal/workflow/domain/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type triggerServiceImpl struct {
	workflowRepo   repository.WorkflowRepository
	execRepo       repository.ExecutionRepository
	channelHandler service.StepHandler
	delayHandler   service.StepHandler
}

func NewTriggerService(
	workflowRepo repository.WorkflowRepository,
	execRepo repository.ExecutionRepository,
	channelHandler service.StepHandler,
	delayHandler service.StepHandler,
) service.TriggerService {
	return &triggerServiceImpl{
		workflowRepo:   workflowRepo,
		execRepo:       execRepo,
		channelHandler: channelHandler,
		delayHandler:   delayHandler,
	}
}

func (s *triggerServiceImpl) Trigger(ctx context.Context, envID uuid.UUID, triggerIdentifier string,
	subscriberKey string, payload map[string]interface{}) error {

	// 1. Find active workflow
	workflow, err := s.workflowRepo.GetByTrigger(ctx, envID, triggerIdentifier)
	if err != nil {
		return fmt.Errorf("workflow not found for trigger %q: %w", triggerIdentifier, err)
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

	if err := s.execRepo.CreateExecution(ctx, exec); err != nil {
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
		if err := s.execRepo.CreateStepExecution(ctx, se); err != nil {
			return fmt.Errorf("failed to create step execution: %w", err)
		}
		stepExecs = append(stepExecs, se)
	}

	// 4. Execute steps starting from the first
	s.executeStepsFrom(ctx, exec, steps, stepExecs, 0)

	return nil
}

// executeStepsFrom processes steps sequentially starting from the given index.
// Stops when a Delay step is encountered (it will be picked up by the worker) or on error.
func (s *triggerServiceImpl) executeStepsFrom(ctx context.Context,
	exec *entity.WorkflowExecution, steps []entity.WorkflowStep,
	stepExecs []*entity.StepExecution, fromIndex int) {

	for i := fromIndex; i < len(steps); i++ {
		step := &steps[i]
		stepExec := stepExecs[i]

		// Update current step pointer
		exec.CurrentStepID = &step.ID
		_ = s.execRepo.UpdateExecution(ctx, exec)

		// Mark step as running
		now := time.Now()
		stepExec.Status = entity.StepStatusRunning
		stepExec.StartedAt = &now
		_ = s.execRepo.UpdateStepExecution(ctx, stepExec)

		var handler service.StepHandler
		switch step.StepType {
		case entity.StepTypeChannel:
			handler = s.channelHandler
		case entity.StepTypeDelay:
			handler = s.delayHandler
		default:
			global.Logger.Error("workflow_trigger: unknown step type", zap.String("stepType", step.StepType), zap.String("executionID", exec.ID.String()))
			s.markStepFailed(ctx, stepExec, exec, "unknown step type: "+step.StepType)
			return
		}

		if err := handler.Execute(ctx, step, exec, stepExec); err != nil {
			global.Logger.Error("workflow_trigger: step execution failed", zap.String("stepID", step.ID.String()), zap.String("executionID", exec.ID.String()), zap.Error(err))
			s.markStepFailed(ctx, stepExec, exec, err.Error())
			return
		}

		// Delay steps set their own status to "scheduled" — stop processing
		if step.StepType == entity.StepTypeDelay {
			global.Logger.Info("Workflow trigger: delay step scheduled, pausing execution " + exec.ID.String())
			return
		}

		// Channel step success — mark completed and continue
		completedAt := time.Now()
		stepExec.Status = entity.StepStatusCompleted
		stepExec.CompletedAt = &completedAt
		if err := s.execRepo.UpdateStepExecution(ctx, stepExec); err != nil {
			global.Logger.Warn("workflow_trigger: failed to update step execution status", zap.String("stepExecID", stepExec.ID.String()), zap.Error(err))
		}
	}

	// All steps completed
	completedAt := time.Now()
	exec.Status = entity.WorkflowExecutionCompleted
	exec.CompletedAt = &completedAt
	if err := s.execRepo.UpdateExecution(ctx, exec); err != nil {
		global.Logger.Warn("workflow_trigger: failed to update execution status", zap.String("executionID", exec.ID.String()), zap.Error(err))
	}

	global.Logger.Info("workflow_trigger: execution completed", zap.String("executionID", exec.ID.String()))
}

func (s *triggerServiceImpl) markStepFailed(ctx context.Context,
	stepExec *entity.StepExecution, exec *entity.WorkflowExecution, reason string) {

	now := time.Now()
	stepExec.Status = entity.StepStatusFailed
	stepExec.CompletedAt = &now
	stepExec.Output = map[string]interface{}{"error": reason}
	if err := s.execRepo.UpdateStepExecution(ctx, stepExec); err != nil {
		global.Logger.Warn("workflow_trigger: failed to mark step as failed", zap.String("stepExecID", stepExec.ID.String()), zap.Error(err))
	}

	exec.Status = entity.WorkflowExecutionFailed
	exec.CompletedAt = &now
	if err := s.execRepo.UpdateExecution(ctx, exec); err != nil {
		global.Logger.Warn("workflow_trigger: failed to mark execution as failed", zap.String("executionID", exec.ID.String()), zap.Error(err))
	}
}
