package adapter

import (
	"context"

	"CONVERDA/internal/notification/domain/repository"
	"CONVERDA/internal/notification/infrastructure/provider"

	"github.com/google/uuid"
)

type DispatcherAdapter struct {
	dispatcher *provider.Dispatcher
}

func NewDispatcherAdapter(dispatcher *provider.Dispatcher) repository.INotificationDispatcher {
	return &DispatcherAdapter{dispatcher: dispatcher}
}

func (a *DispatcherAdapter) Dispatch(ctx context.Context, envID uuid.UUID, channel string, recipient string, subject string, body string, data map[string]string) error {
	return a.dispatcher.Dispatch(ctx, envID, channel, recipient, subject, body, data)
}
