package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"CONVERDA/global"
	"CONVERDA/internal/workflow/application/service"
	"CONVERDA/internal/workflow/domain/model/entity"
	"CONVERDA/internal/workflow/domain/repository"
)

const (
	defaultPollInterval = 30 * time.Second
	defaultBatchSize    = 50
)

// WorkflowExecutor polls for scheduled step executions and processes them.
// It handles Delay steps that have reached their scheduled_at time.
type WorkflowExecutor struct {
	execRepo       repository.ExecutionRepository
	workflowRepo   repository.WorkflowRepository
	channelHandler service.StepHandler
	delayHandler   service.StepHandler
	digestHandler  service.StepHandler
	interval       time.Duration
	batchSize      int
}

func NewWorkflowExecutor(
	execRepo repository.ExecutionRepository,
	workflowRepo repository.WorkflowRepository,
	channelHandler service.StepHandler,
	delayHandler service.StepHandler,
	digestHandler service.StepHandler,
) *WorkflowExecutor {
	return &WorkflowExecutor{
		execRepo:       execRepo,
		workflowRepo:   workflowRepo,
		channelHandler: channelHandler,
		delayHandler:   delayHandler,
		digestHandler:  digestHandler,
		interval:       defaultPollInterval,
		batchSize:      defaultBatchSize,
	}
}

// Run starts the polling loop. Blocks until context is cancelled.
func (w *WorkflowExecutor) Run(ctx context.Context) {
	global.Logger.Info("WorkflowExecutor: started (polling every " + w.interval.String() + ")")

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			global.Logger.Info("WorkflowExecutor: shutting down")
			return
		case <-ticker.C:
			w.poll(ctx)
			w.pollDigest(ctx)
		}
	}
}

func (w *WorkflowExecutor) poll(ctx context.Context) {
	// Find step executions that are scheduled and due
	stepExecs, err := w.execRepo.GetPendingScheduledSteps(ctx, w.batchSize)
	if err != nil {
		global.Logger.Error("WorkflowExecutor: failed to get pending steps: " + err.Error())
		return
	}

	if len(stepExecs) == 0 {
		return
	}

	global.Logger.Info("WorkflowExecutor: processing " + string(rune('0'+len(stepExecs))) + " scheduled steps")

	for _, stepExec := range stepExecs {
		w.processScheduledStep(ctx, stepExec)
	}
}

func (w *WorkflowExecutor) processScheduledStep(ctx context.Context, stepExec *entity.StepExecution) {
	// Get the parent execution
	exec, err := w.execRepo.GetExecution(ctx, stepExec.ExecutionID)
	if err != nil {
		global.Logger.Error("WorkflowExecutor: failed to get execution " + stepExec.ExecutionID.String() + ": " + err.Error())
		return
	}

	// Get the workflow with steps
	workflow, err := w.workflowRepo.GetByID(ctx, exec.WorkflowID)
	if err != nil {
		global.Logger.Error("WorkflowExecutor: failed to get workflow " + exec.WorkflowID.String() + ": " + err.Error())
		return
	}

	// The scheduled step (Delay) is now due — mark it completed
	now := time.Now()
	stepExec.Status = entity.StepStatusCompleted
	stepExec.CompletedAt = &now
	if err := w.execRepo.UpdateStepExecution(ctx, stepExec); err != nil {
		global.Logger.Error("WorkflowExecutor: failed to update step execution: " + err.Error())
		return
	}

	// Find the next step after this one
	steps := make([]entity.WorkflowStep, len(workflow.Steps))
	copy(steps, workflow.Steps)
	sort.Slice(steps, func(i, j int) bool {
		return steps[i].Order < steps[j].Order
	})

	// Find current step index
	currentIdx := -1
	for i, s := range steps {
		if s.ID == stepExec.StepID {
			currentIdx = i
			break
		}
	}

	if currentIdx == -1 || currentIdx >= len(steps)-1 {
		// No more steps — workflow complete
		completedAt := time.Now()
		exec.Status = entity.WorkflowExecutionCompleted
		exec.CompletedAt = &completedAt
		_ = w.execRepo.UpdateExecution(ctx, exec)
		global.Logger.Info("WorkflowExecutor: execution " + exec.ID.String() + " completed")
		return
	}

	// Get all step executions for this workflow execution
	allStepExecs, err := w.execRepo.GetStepExecutionsByExecution(ctx, exec.ID)
	if err != nil {
		global.Logger.Error("WorkflowExecutor: failed to get step executions: " + err.Error())
		return
	}

	// Build step exec map
	stepExecMap := make(map[string]*entity.StepExecution)
	for _, se := range allStepExecs {
		stepExecMap[se.StepID.String()] = se
	}

	// Process remaining steps starting from the next one
	for i := currentIdx + 1; i < len(steps); i++ {
		step := &steps[i]
		se := stepExecMap[step.ID.String()]
		if se == nil {
			global.Logger.Error("WorkflowExecutor: step execution not found for step " + step.ID.String())
			return
		}

		// Update current step pointer
		exec.CurrentStepID = &step.ID
		_ = w.execRepo.UpdateExecution(ctx, exec)

		// Mark step as running
		startTime := time.Now()
		se.Status = entity.StepStatusRunning
		se.StartedAt = &startTime
		_ = w.execRepo.UpdateStepExecution(ctx, se)

		var handler service.StepHandler
		switch step.StepType {
		case entity.StepTypeChannel:
			handler = w.channelHandler
		case entity.StepTypeDelay:
			handler = w.delayHandler
		case entity.StepTypeDigest:
			handler = w.digestHandler
		default:
			global.Logger.Error("WorkflowExecutor: unknown step type " + step.StepType)
			w.markFailed(ctx, se, exec, "unknown step type: "+step.StepType)
			return
		}

		if err := handler.Execute(ctx, step, exec, se); err != nil {
			global.Logger.Error("WorkflowExecutor: step " + step.ID.String() + " failed: " + err.Error())
			w.markFailed(ctx, se, exec, err.Error())
			return
		}

		// Delay/Digest step scheduled itself — stop here (will be picked up next poll)
		if step.StepType == entity.StepTypeDelay || step.StepType == entity.StepTypeDigest {
			return
		}

		// Channel step completed
		completedTime := time.Now()
		se.Status = entity.StepStatusCompleted
		se.CompletedAt = &completedTime
		_ = w.execRepo.UpdateStepExecution(ctx, se)
	}

	// All steps done
	completedAt := time.Now()
	exec.Status = entity.WorkflowExecutionCompleted
	exec.CompletedAt = &completedAt
	_ = w.execRepo.UpdateExecution(ctx, exec)
	global.Logger.Info("WorkflowExecutor: execution " + exec.ID.String() + " completed")
}

// pollDigest flushes digest steps whose window has expired.
func (w *WorkflowExecutor) pollDigest(ctx context.Context) {
	stepExecs, err := w.execRepo.GetDigestingSteps(ctx, w.batchSize)
	if err != nil {
		global.Logger.Error("WorkflowExecutor: failed to get digesting steps: " + err.Error())
		return
	}

	for _, stepExec := range stepExecs {
		w.flushDigestStep(ctx, stepExec)
	}
}

func (w *WorkflowExecutor) flushDigestStep(ctx context.Context, stepExec *entity.StepExecution) {
	exec, err := w.execRepo.GetExecution(ctx, stepExec.ExecutionID)
	if err != nil {
		global.Logger.Error("WorkflowExecutor: failed to get execution for digest flush: " + err.Error())
		return
	}

	// Flush buffered events
	events, err := w.execRepo.FlushDigestEvents(ctx, stepExec.StepID, exec.ID)
	if err != nil {
		global.Logger.Error("WorkflowExecutor: failed to flush digest events: " + err.Error())
		return
	}

	// Merge events into execution payload
	digestPayloads := make([]map[string]interface{}, 0, len(events))
	for _, ev := range events {
		digestPayloads = append(digestPayloads, ev.Payload)
	}
	digestJSON, _ := json.Marshal(digestPayloads)
	if exec.TriggerPayload == nil {
		exec.TriggerPayload = make(map[string]interface{})
	}
	exec.TriggerPayload["digest_events"] = json.RawMessage(digestJSON)
	exec.TriggerPayload["digest_count"] = len(events)
	_ = w.execRepo.UpdateExecution(ctx, exec)

	// Mark digest step completed
	now := time.Now()
	stepExec.Status = entity.StepStatusCompleted
	stepExec.CompletedAt = &now
	stepExec.Output = map[string]interface{}{"events_count": len(events)}
	_ = w.execRepo.UpdateStepExecution(ctx, stepExec)

	global.Logger.Info(fmt.Sprintf("WorkflowExecutor: digest step %s flushed %d events", stepExec.StepID, len(events)))

	// Continue to next step (reuse processScheduledStep logic)
	w.processScheduledStep(ctx, stepExec)
}

func (w *WorkflowExecutor) markFailed(ctx context.Context,
	stepExec *entity.StepExecution, exec *entity.WorkflowExecution, reason string) {

	now := time.Now()
	stepExec.Status = entity.StepStatusFailed
	stepExec.CompletedAt = &now
	stepExec.Output = map[string]interface{}{"error": reason}
	_ = w.execRepo.UpdateStepExecution(ctx, stepExec)

	exec.Status = entity.WorkflowExecutionFailed
	exec.CompletedAt = &now
	_ = w.execRepo.UpdateExecution(ctx, exec)
}
