package impl

import (
	"context"
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
		return nil, err
	}
	return s.providerRepo.Create(ctx, provider)
}

func (s *providerServiceImpl) GetProvider(ctx context.Context, id uuid.UUID) (*entity.Provider, error) {
	return s.providerRepo.GetByID(ctx, id)
}

func (s *providerServiceImpl) ListProviders(ctx context.Context, appID uuid.UUID) ([]*entity.Provider, error) {
	return s.providerRepo.ListByApp(ctx, appID)
}

func (s *providerServiceImpl) GetActiveProvider(ctx context.Context, envID uuid.UUID, providerType string) (*entity.Provider, error) {
	return s.providerRepo.GetByTypeAndEnv(ctx, envID, providerType)
}

func (s *providerServiceImpl) UpdateProvider(ctx context.Context, provider *entity.Provider) error {
	provider.UpdatedAt = time.Now()
	return s.providerRepo.Update(ctx, provider)
}

func (s *providerServiceImpl) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	return s.providerRepo.Delete(ctx, id)
}

func (s *providerServiceImpl) ToggleProvider(ctx context.Context, id uuid.UUID, active bool) error {
	provider, err := s.providerRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	provider.IsActive = active
	provider.UpdatedAt = time.Now()
	return s.providerRepo.Update(ctx, provider)
}
