package service

import (
	"context"

	"github.com/google/uuid"
)

// TriggerService handles workflow triggering from external events.
type TriggerService interface {
	// Trigger looks up an active workflow by (envID, triggerIdentifier),
	// creates execution records, and starts processing the first step.
	Trigger(ctx context.Context, envID uuid.UUID, triggerIdentifier string,
		subscriberKey string, payload map[string]interface{}) error
}
