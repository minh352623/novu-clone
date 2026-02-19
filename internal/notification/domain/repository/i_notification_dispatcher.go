package repository

import (
	"context"

	"github.com/google/uuid"
)

// INotificationDispatcher defines the port for dispatching notifications across various channels.
type INotificationDispatcher interface {
	Dispatch(ctx context.Context, envID uuid.UUID, channel string, recipient string, subject string, body string, data map[string]string) error
}
