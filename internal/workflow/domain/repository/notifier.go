package repository

import (
	"context"

	"github.com/google/uuid"
)

// NotificationRequest defines the data needed to send a notification from Workflow module
type NotificationRequest struct {
	TenantID      uuid.UUID
	EnvironmentID uuid.UUID
	Recipient     string
	TemplateCode  string
	Channel       string
	Language      string
	Data          map[string]interface{}
}

// Notifier defines the interface for sending notifications from the Workflow module
// following the Interface+Adapter pattern (GOLANG_BEST_PRACTICES.md Section 11)
type Notifier interface {
	Send(ctx context.Context, req NotificationRequest) error
}
