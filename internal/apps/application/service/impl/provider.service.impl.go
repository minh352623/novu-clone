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

type providerServiceImpl struct {
	providerRepo repository.ProviderRepository
}

func NewProviderService(providerRepo repository.ProviderRepository) service.ProviderService {
	return &providerServiceImpl{providerRepo: providerRepo}
}

func (s *providerServiceImpl) CreateProvider(ctx context.Context, tenantID, appID, envID uuid.UUID, providerType, providerName string, config map[string]interface{}) (*entity.Provider, error) {
	provider, err := entity.NewProvider(tenantID, appID, envID, providerType, providerName, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider entity: %w", err)
	}
	created, err := s.providerRepo.Create(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}
	return created, nil
}

func (s *providerServiceImpl) GetProvider(ctx context.Context, id uuid.UUID) (*entity.Provider, error) {
	provider, err := s.providerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider %s: %w", id, err)
	}
	return provider, nil
}

func (s *providerServiceImpl) ListProviders(ctx context.Context, appID uuid.UUID) ([]*entity.Provider, error) {
	providers, err := s.providerRepo.ListByApp(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to list providers for app %s: %w", appID, err)
	}
	return providers, nil
}

func (s *providerServiceImpl) GetActiveProvider(ctx context.Context, envID uuid.UUID, providerType string) (*entity.Provider, error) {
	provider, err := s.providerRepo.GetByTypeAndEnv(ctx, envID, providerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get active provider for env %s type %s: %w", envID, providerType, err)
	}
	return provider, nil
}

func (s *providerServiceImpl) UpdateProvider(ctx context.Context, provider *entity.Provider) error {
	provider.UpdatedAt = time.Now()
	if err := s.providerRepo.Update(ctx, provider); err != nil {
		return fmt.Errorf("failed to update provider %s: %w", provider.ID, err)
	}
	return nil
}

func (s *providerServiceImpl) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	if err := s.providerRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete provider %s: %w", id, err)
	}
	return nil
}

func (s *providerServiceImpl) ToggleProvider(ctx context.Context, id uuid.UUID, active bool) error {
	provider, err := s.providerRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}
	provider.IsActive = active
	provider.UpdatedAt = time.Now()
	return s.providerRepo.Update(ctx, provider)
}
