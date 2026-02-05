package impl

import (
	"context"
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
		return nil, err
	}
	return s.webhookRepo.Create(ctx, webhook)
}

func (s *webhookServiceImpl) GetWebhook(ctx context.Context, id uuid.UUID) (*entity.Webhook, error) {
	return s.webhookRepo.GetByID(ctx, id)
}

func (s *webhookServiceImpl) ListWebhooks(ctx context.Context, appID uuid.UUID) ([]*entity.Webhook, error) {
	return s.webhookRepo.ListByApp(ctx, appID)
}

func (s *webhookServiceImpl) UpdateWebhook(ctx context.Context, webhook *entity.Webhook) error {
	webhook.UpdatedAt = time.Now()
	return s.webhookRepo.Update(ctx, webhook)
}

func (s *webhookServiceImpl) DeleteWebhook(ctx context.Context, id uuid.UUID) error {
	return s.webhookRepo.Delete(ctx, id)
}

func (s *webhookServiceImpl) ToggleWebhook(ctx context.Context, id uuid.UUID, active bool) error {
	webhook, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	webhook.IsActive = active
	webhook.UpdatedAt = time.Now()
	return s.webhookRepo.Update(ctx, webhook)
}
