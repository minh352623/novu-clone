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
)

type appServiceImpl struct {
	appRepo       repository.AppRepository
	envRepo       repository.EnvironmentRepository
	systemEnvServ service.SystemEnvironmentService
	apiKeyServ    service.APIKeyService
	uow           repository.AppsUnitOfWork
}

func NewAppService(
	appRepo repository.AppRepository,
	envRepo repository.EnvironmentRepository,
	systemEnvServ service.SystemEnvironmentService,
	apiKeyServ service.APIKeyService,
	uow repository.AppsUnitOfWork,
) service.AppService {
	return &appServiceImpl{
		appRepo:       appRepo,
		envRepo:       envRepo,
		systemEnvServ: systemEnvServ,
		apiKeyServ:    apiKeyServ,
		uow:           uow,
	}
}

func (s *appServiceImpl) CreateApp(ctx context.Context, tenantID uuid.UUID, name string, description *string, slaThreshold *int) (*entity.App, error) {
	var createdApp *entity.App

	// Fetch system environments first (Read-only, before transaction is best, or inside if needed)
	systemEnvs, err := s.systemEnvServ.ListAll(ctx)
	if err != nil {
		global.Logger.Warn("auto-provisioning: failed to fetch system environments", "error", err)
		// Should we fail or continue? Business decision. Existing code continued with warning.
		// However, for atomic provisioning, we probably want them.
	}

	err = s.uow.Execute(ctx, func(tx repository.AppsTxRepository) error {
		app, err := entity.NewApp(tenantID, name, description)
		if err != nil {
			return fmt.Errorf("failed to initialize app: %w", err)
		}
		if slaThreshold != nil {
			app.SLAThresholdSeconds = *slaThreshold
		}

		createdApp, err = tx.Apps().Create(ctx, app)
		if err != nil {
			return fmt.Errorf("failed to create app: %w", err)
		}

		// Auto-create default environments
		for _, se := range systemEnvs {
			env, err := entity.NewEnvironment(createdApp.ID, se.Code)
			if err != nil {
				continue
			}
			createdEnv, err := tx.Environments().Create(ctx, env)
			if err != nil {
				return fmt.Errorf("failed to auto-provision environment %s: %w", se.Code, err)
			}

			// Auto-generate default API Key for the environment
			_, apiKey, err := entity.GenerateAPIKey(createdApp.ID, createdEnv.ID, se.Code, "Default Key")
			if err != nil {
				return fmt.Errorf("failed to generate default API key for %s: %w", se.Code, err)
			}

			if _, err := tx.APIKeys().Create(ctx, apiKey); err != nil {
				return fmt.Errorf("failed to save default API key for %s: %w", se.Code, err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
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
