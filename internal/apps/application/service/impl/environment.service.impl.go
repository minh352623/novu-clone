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

type environmentServiceImpl struct {
	envRepo repository.EnvironmentRepository
	uow     repository.AppsUnitOfWork
}

func NewEnvironmentService(envRepo repository.EnvironmentRepository, uow repository.AppsUnitOfWork) service.EnvironmentService {
	return &environmentServiceImpl{
		envRepo: envRepo,
		uow:     uow,
	}
}

func (s *environmentServiceImpl) CreateEnvironment(ctx context.Context, appID uuid.UUID, code string, slaThreshold *int) (*entity.Environment, error) {
	var created *entity.Environment

	err := s.uow.Execute(ctx, func(tx repository.AppsTxRepository) error {
		// Check if exists
		existing, _ := tx.Environments().GetByAppAndCode(ctx, appID, code)
		if existing != nil {
			return entity.ErrDuplicateEnvironment
		}

		env, err := entity.NewEnvironment(appID, code)
		if err != nil {
			return err
		}
		if slaThreshold != nil {
			env.SLAThresholdSeconds = *slaThreshold
		}

		created, err = tx.Environments().Create(ctx, env)
		if err != nil {
			return fmt.Errorf("failed to create environment: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *environmentServiceImpl) ListEnvironments(ctx context.Context, appID uuid.UUID) ([]*entity.Environment, error) {
	envs, err := s.envRepo.ListByApp(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to list environments for app %s: %w", appID, err)
	}
	return envs, nil
}

func (s *environmentServiceImpl) GetEnvironment(ctx context.Context, appID uuid.UUID, code string) (*entity.Environment, error) {
	env, err := s.envRepo.GetByAppAndCode(ctx, appID, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment by code %s: %w", code, err)
	}
	if env == nil {
		return nil, entity.ErrEnvironmentNotFound
	}
	return env, nil
}

func (s *environmentServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Environment, error) {
	env, err := s.envRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment: %w", err)
	}
	if env == nil {
		return nil, entity.ErrEnvironmentNotFound
	}
	return env, nil
}

func (s *environmentServiceImpl) VerifyAPIKey(ctx context.Context, apiKey string) (*entity.Environment, error) {
	env, err := s.envRepo.GetByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to verify api key: %w", err)
	}
	if env == nil {
		return nil, entity.ErrEnvironmentNotFound
	}
	return env, nil
}

func (s *environmentServiceImpl) UpdateEnvironment(ctx context.Context, env *entity.Environment) error {
	env.UpdatedAt = time.Now()
	if err := s.envRepo.Update(ctx, env); err != nil {
		return fmt.Errorf("failed to update environment %s: %w", env.ID, err)
	}
	return nil
}
