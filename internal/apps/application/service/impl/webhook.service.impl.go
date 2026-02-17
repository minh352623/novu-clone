package impl

import (
	"context"
	"fmt"
	"time"

	"CONVERDA/internal/apps/application/service"
	"CONVERDA/internal/apps/domain/model/entity"
	"CONVERDA/internal/apps/domain/repository"

	"github.com/google/uuid"
)

type webhookServiceImpl struct {
	webhookRepo repository.WebhookRepository
}

func NewWebhookService(webhookRepo repository.WebhookRepository) service.WebhookService {
	return &webhookServiceImpl{webhookRepo: webhookRepo}
}

func (s *webhookServiceImpl) CreateWebhook(ctx context.Context, tenantID, appID, envID uuid.UUID, url string, events []string, desc *string) (*entity.Webhook, error) {
	webhook, err := entity.NewWebhook(tenantID, appID, envID, url, events, desc)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook entity: %w", err)
	}
	created, err := s.webhookRepo.Create(ctx, webhook)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}
	return created, nil
}

func (s *webhookServiceImpl) GetWebhook(ctx context.Context, id uuid.UUID) (*entity.Webhook, error) {
	webhook, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook %s: %w", id, err)
	}
	return webhook, nil
}

func (s *webhookServiceImpl) ListWebhooks(ctx context.Context, appID uuid.UUID) ([]*entity.Webhook, error) {
	webhooks, err := s.webhookRepo.ListByApp(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks for app %s: %w", appID, err)
	}
	return webhooks, nil
}

func (s *webhookServiceImpl) UpdateWebhook(ctx context.Context, webhook *entity.Webhook) error {
	webhook.UpdatedAt = time.Now()
	if err := s.webhookRepo.Update(ctx, webhook); err != nil {
		return fmt.Errorf("failed to update webhook %s: %w", webhook.ID, err)
	}
	return nil
}

func (s *webhookServiceImpl) DeleteWebhook(ctx context.Context, id uuid.UUID) error {
	if err := s.webhookRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete webhook %s: %w", id, err)
	}
	return nil
}

func (s *webhookServiceImpl) ToggleWebhook(ctx context.Context, id uuid.UUID, active bool) error {
	webhook, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}
	webhook.IsActive = active
	webhook.UpdatedAt = time.Now()
	return s.webhookRepo.Update(ctx, webhook)
}
