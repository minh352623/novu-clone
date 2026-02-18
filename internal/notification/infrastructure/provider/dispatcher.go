package provider

import (
	"context"
	"fmt"

	"CONVERDA/global"

	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/domain/repository"

	"github.com/google/uuid"
)

type EmailProvider interface {
	Send(ctx context.Context, to, subject, body string) error
}

type PushProvider interface {
	Send(ctx context.Context, to, subject, body string, data map[string]string) error
}

type Dispatcher struct {
	configRepo repository.ProviderConfigRepository
	emailProv  EmailProvider // For testing injection
	pushProv   PushProvider  // For testing injection
}

func NewDispatcher(configRepo repository.ProviderConfigRepository) *Dispatcher {
	return &Dispatcher{configRepo: configRepo}
}

func (d *Dispatcher) SetEmailProvider(p EmailProvider) {
	d.emailProv = p
}
func (d *Dispatcher) SetPushProvider(p PushProvider) {
	d.pushProv = p
}

func (d *Dispatcher) Dispatch(ctx context.Context, envID uuid.UUID, channel, recipient, subject, body string, data map[string]string) error {
	// 1. Get Provider Config
	config, err := d.configRepo.GetActive(ctx, envID, channel)
	if err != nil {
		// Fallback to default or error?
		// For now, if no config, we log and simulate success (or error)
		global.Logger.Warn("Dispatcher: no provider config", "channel", channel, "envID", envID.String(), "error", err)
		return fmt.Errorf("provider not configured for channel %s", channel)
	}

	// 2. Dispatch based on type
	switch config.Type {
	case "smtp":
		return d.sendEmail(ctx, config, recipient, subject, body)
	case "fcm":
		return d.sendPush(ctx, config, recipient, subject, body, data)
	default:
		return fmt.Errorf("unsupported provider type: %s", config.Type)
	}
}

func (d *Dispatcher) sendEmail(ctx context.Context, config *entity.ProviderConfig, recipient, subject, body string) error {
	if d.emailProv != nil {
		return d.emailProv.Send(ctx, recipient, subject, body)
	}

	p, err := NewSmtpProvider(config.Config)
	if err != nil {
		return fmt.Errorf("failed to init smtp provider: %w", err)
	}

	global.Logger.Info("SMTP: dispatching email", "recipient", recipient)
	return p.Send(ctx, recipient, subject, body)
}

func (d *Dispatcher) sendPush(ctx context.Context, config *entity.ProviderConfig, recipient, subject, body string, data map[string]string) error {
	if d.pushProv != nil {
		return d.pushProv.Send(ctx, recipient, subject, body, data)
	}

	p, err := NewFcmProvider(config.Config)
	if err != nil {
		return fmt.Errorf("failed to init fcm provider: %w", err)
	}

	return p.Send(ctx, recipient, subject, body, data)
}
