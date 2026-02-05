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

type appServiceImpl struct {
	appRepo repository.AppRepository
	envRepo repository.EnvironmentRepository
}

func NewAppService(appRepo repository.AppRepository, envRepo repository.EnvironmentRepository) service.AppService {
	return &appServiceImpl{
		appRepo: appRepo,
		envRepo: envRepo,
	}
}

func (s *appServiceImpl) CreateApp(ctx context.Context, tenantID uuid.UUID, name string, description *string) (*entity.App, error) {
	app, err := entity.NewApp(tenantID, name, description)
	if err != nil {
		return nil, err
	}

	createdApp, err := s.appRepo.Create(ctx, app)
	if err != nil {
		return nil, fmt.Errorf("failed to create app: %w", err)
	}

	// Auto-create default environments
	defaultEnvs := []string{"development", "staging", "production"}
	for _, code := range defaultEnvs {
		env, err := entity.NewEnvironment(createdApp.ID, code)
		if err == nil {
			_, _ = s.envRepo.Create(ctx, env)
		}
	}

	return createdApp, nil
}

func (s *appServiceImpl) GetApp(ctx context.Context, appID uuid.UUID) (*entity.App, error) {
	return s.appRepo.GetByID(ctx, appID)
}

func (s *appServiceImpl) ListApps(ctx context.Context, tenantID uuid.UUID) ([]*entity.App, error) {
	return s.appRepo.ListByTenant(ctx, tenantID)
}

func (s *appServiceImpl) UpdateApp(ctx context.Context, app *entity.App) error {
	app.UpdatedAt = time.Now()
	return s.appRepo.Update(ctx, app)
}

func (s *appServiceImpl) DeleteApp(ctx context.Context, appID uuid.UUID) error {
	return s.appRepo.Delete(ctx, appID)
}
