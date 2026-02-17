package impl

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"CONVERDA/internal/workflow/domain/model/entity"
	"CONVERDA/internal/workflow/domain/repository"
)

// DelayHandler schedules the step execution for a future time.
type DelayHandler struct {
	execRepo repository.ExecutionRepository
}

func NewDelayHandler(execRepo repository.ExecutionRepository) *DelayHandler {
	return &DelayHandler{execRepo: execRepo}
}

func (h *DelayHandler) Execute(ctx context.Context, step *entity.WorkflowStep,
	exec *entity.WorkflowExecution, stepExec *entity.StepExecution) error {

	duration, err := parseDuration(step.Config)
	if err != nil {
		return fmt.Errorf("delay step %s: %w", step.ID, err)
	}

	scheduledAt := time.Now().Add(duration)
	stepExec.ScheduledAt = &scheduledAt
	stepExec.Status = entity.StepStatusScheduled

	return h.execRepo.UpdateStepExecution(ctx, stepExec)
}

// parseDuration reads delay duration from step config.
// Supports: "duration": "1h", "duration": "30m", "duration_seconds": 3600
func parseDuration(config map[string]interface{}) (time.Duration, error) {
	// Try "duration" string first (e.g., "1h", "30m", "5s")
	if durationStr, ok := config["duration"].(string); ok && durationStr != "" {
		return time.ParseDuration(durationStr)
	}

	// Try "duration_seconds" numeric
	switch v := config["duration_seconds"].(type) {
	case float64:
		return time.Duration(int64(v)) * time.Second, nil
	case int:
		return time.Duration(v) * time.Second, nil
	case string:
		sec, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("invalid duration_seconds: %s", v)
		}
		return time.Duration(sec) * time.Second, nil
	}

	return 0, fmt.Errorf("missing duration or duration_seconds in config")
}
