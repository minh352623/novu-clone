package impl

import (
	"context"
	"fmt"
	"time"

	"CONVERDA/internal/workflow/domain/model/entity"
	"CONVERDA/internal/workflow/domain/repository"

	"github.com/google/uuid"
)

// DigestHandler buffers trigger events during a configurable time window.
// When the window expires (handled by WorkflowExecutor), buffered events are
// flushed and merged into the execution payload for the next step.
type DigestHandler struct {
	execRepo repository.ExecutionRepository
}

func NewDigestHandler(execRepo repository.ExecutionRepository) *DigestHandler {
	return &DigestHandler{execRepo: execRepo}
}

func (h *DigestHandler) Execute(ctx context.Context, step *entity.WorkflowStep,
	exec *entity.WorkflowExecution, stepExec *entity.StepExecution) error {

	window, err := parseWindow(step.Config)
	if err != nil {
		return fmt.Errorf("digest step %s: %w", step.ID, err)
	}

	// Buffer the current trigger payload as a digest event
	event := &entity.DigestEvent{
		ID:            uuid.New(),
		StepID:        step.ID,
		ExecutionID:   exec.ID,
		SubscriberKey: exec.SubscriberKey,
		Payload:       exec.TriggerPayload,
		CreatedAt:     time.Now(),
	}

	if err := h.execRepo.BufferDigestEvent(ctx, event); err != nil {
		return fmt.Errorf("digest step %s: failed to buffer event: %w", step.ID, err)
	}

	// Schedule the digest window end
	scheduledAt := time.Now().Add(window)
	stepExec.ScheduledAt = &scheduledAt
	stepExec.Status = entity.StepStatusDigesting

	return h.execRepo.UpdateStepExecution(ctx, stepExec)
}

// parseWindow reads the digest window duration from step config.
// Supports: "window": "1h", "window": "30m", "window_seconds": 3600
func parseWindow(config map[string]interface{}) (time.Duration, error) {
	// Try "window" string first
	if windowStr, ok := config["window"].(string); ok && windowStr != "" {
		return time.ParseDuration(windowStr)
	}

	// Try "window_seconds" numeric
	switch v := config["window_seconds"].(type) {
	case float64:
		return time.Duration(int64(v)) * time.Second, nil
	case int:
		return time.Duration(v) * time.Second, nil
	}

	return 0, fmt.Errorf("missing window or window_seconds in config")
}
