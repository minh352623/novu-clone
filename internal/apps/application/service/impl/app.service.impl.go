package impl

import (
	"context"
	"fmt"

	"CONVERDA/global"
	"time"

	"CONVERDA/internal/apps/application/service"
	"CONVERDA/internal/apps/domain/model/entity"
	"CONVERDA/internal/apps/domain/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type appServiceImpl struct {
	appRepo       repository.AppRepository
	envRepo       repository.EnvironmentRepository
	systemEnvServ service.SystemEnvironmentService
	apiKeyServ    service.APIKeyService
}

func NewAppService(
	appRepo repository.AppRepository,
	envRepo repository.EnvironmentRepository,
	systemEnvServ service.SystemEnvironmentService,
	apiKeyServ service.APIKeyService,
) service.AppService {
	return &appServiceImpl{
		appRepo:       appRepo,
		envRepo:       envRepo,
		systemEnvServ: systemEnvServ,
		apiKeyServ:    apiKeyServ,
	}
}

func (s *appServiceImpl) CreateApp(ctx context.Context, tenantID uuid.UUID, name string, description *string, slaThreshold *int) (*entity.App, error) {
	app, err := entity.NewApp(tenantID, name, description)
	if slaThreshold != nil {
		app.SLAThresholdSeconds = *slaThreshold
	}

	createdApp, err := s.appRepo.Create(ctx, app)
	if err != nil {
		return nil, fmt.Errorf("failed to create app: %w", err)
	}

	// Auto-create default environments from database
	systemEnvs, err := s.systemEnvServ.ListAll(ctx)
	if err == nil {
		for _, se := range systemEnvs {
			env, err := entity.NewEnvironment(createdApp.ID, se.Code)
			if err != nil {
				continue
			}
			createdEnv, err := s.envRepo.Create(ctx, env)
			if err != nil {
				global.Logger.Warn("auto-provisioning: failed to create environment", zap.String("code", se.Code), zap.Error(err))
				continue
			}

			// Auto-generate default API Key for the environment
			_, _, err = s.apiKeyServ.GenerateKey(ctx, createdEnv.ID, "Default Key")
			if err != nil {
				global.Logger.Warn("auto-provisioning: failed to generate API key", zap.String("code", se.Code), zap.Error(err))
			}
		}
	} else {
		global.Logger.Warn("auto-provisioning: failed to fetch system environments", zap.Error(err))
	}

	return createdApp, nil
}

func (s *appServiceImpl) GetApp(ctx context.Context, appID uuid.UUID) (*entity.App, error) {
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app by id %s: %w", appID, err)
	}
	return app, nil
}

func (s *appServiceImpl) ListApps(ctx context.Context, tenantID uuid.UUID) ([]*entity.App, error) {
	apps, err := s.appRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list apps for tenant %s: %w", tenantID, err)
	}
	return apps, nil
}

func (s *appServiceImpl) UpdateApp(ctx context.Context, app *entity.App) error {
	app.UpdatedAt = time.Now()
	if err := s.appRepo.Update(ctx, app); err != nil {
		return fmt.Errorf("failed to update app %s: %w", app.ID, err)
	}
	return nil
}

func (s *appServiceImpl) DeleteApp(ctx context.Context, appID uuid.UUID) error {
	if err := s.appRepo.Delete(ctx, appID); err != nil {
		return fmt.Errorf("failed to delete app %s: %w", appID, err)
	}
	return nil
}
