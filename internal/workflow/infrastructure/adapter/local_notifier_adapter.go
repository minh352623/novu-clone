package adapter

import (
	"context"

	notifService "CONVERDA/internal/notification/application/service"
	"CONVERDA/internal/workflow/domain/repository"
)

type localNotifierAdapter struct {
	notifSvc notifService.NotificationService
}

// NewLocalNotifierAdapter creates a new local adapter for Workflow Notifier
func NewLocalNotifierAdapter(notifSvc notifService.NotificationService) repository.Notifier {
	return &localNotifierAdapter{notifSvc: notifSvc}
}

func (a *localNotifierAdapter) Send(ctx context.Context, req repository.NotificationRequest) error {
	_, err := a.notifSvc.Send(ctx, notifService.SendRequest{
		TenantID:      req.TenantID.String(),
		EnvironmentID: req.EnvironmentID.String(),
		TemplateCode:  req.TemplateCode,
		Recipient:     req.Recipient,
		Channel:       req.Channel,
		Language:      req.Language,
		Data:          req.Data,
	})
	return err
}
