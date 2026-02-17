package adapter

import (
	"context"

	appsRepo "CONVERDA/internal/apps/domain/repository"
	"CONVERDA/internal/messaging/domain/repository"

	"github.com/google/uuid"
)

type localAppAdapter struct {
	appRepo appsRepo.AppRepository
	envRepo appsRepo.EnvironmentRepository
}

// NewLocalAppAdapter creates a new local adapter for AppReader
func NewLocalAppAdapter(appRepo appsRepo.AppRepository, envRepo appsRepo.EnvironmentRepository) repository.AppReader {
	return &localAppAdapter{
		appRepo: appRepo,
		envRepo: envRepo,
	}
}

func (a *localAppAdapter) GetApp(ctx context.Context, id uuid.UUID) (*repository.AppInfo, error) {
	app, err := a.appRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, nil
	}
	return &repository.AppInfo{
		ID:                  app.ID,
		TenantID:            app.TenantID,
		Name:                app.Name,
		SLAThresholdSeconds: app.SLAThresholdSeconds,
	}, nil
}

func (a *localAppAdapter) GetEnvironment(ctx context.Context, id uuid.UUID) (*repository.EnvInfo, error) {
	env, err := a.envRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if env == nil {
		return nil, nil
	}
	return &repository.EnvInfo{
		ID:                  env.ID,
		AppID:               env.AppID,
		Code:                env.EnvironmentCode,
		SLAThresholdSeconds: env.SLAThresholdSeconds,
	}, nil
}
