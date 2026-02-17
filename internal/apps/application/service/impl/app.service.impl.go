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
				fmt.Printf("Warning: failed to auto-create environment %s: %v\n", se.Code, err)
				continue
			}

			// Auto-generate default API Key for the environment
			_, _, err = s.apiKeyServ.GenerateKey(ctx, createdEnv.ID, "Default Key")
			if err != nil {
				fmt.Printf("Warning: failed to auto-generate default API Key for environment %s: %v\n", se.Code, err)
			}
		}
	} else {
		fmt.Printf("Warning: failed to fetch system environments for auto-provisioning: %v\n", err)
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
