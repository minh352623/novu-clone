package impl

import (
	"context"
	"fmt"

	"CONVERDA/internal/apps/application/service"
	"CONVERDA/internal/apps/domain/model/entity"
	"CONVERDA/internal/apps/domain/repository"

	"github.com/google/uuid"
)

type environmentServiceImpl struct {
	envRepo repository.EnvironmentRepository
}

func NewEnvironmentService(envRepo repository.EnvironmentRepository) service.EnvironmentService {
	return &environmentServiceImpl{
		envRepo: envRepo,
	}
}

func (s *environmentServiceImpl) CreateEnvironment(ctx context.Context, appID uuid.UUID, code string) (*entity.Environment, error) {
	// Check if exists
	existing, _ := s.envRepo.GetByAppAndCode(ctx, appID, code)
	if existing != nil {
		return nil, entity.ErrDuplicateEnvironment
	}

	env, err := entity.NewEnvironment(appID, code)
	if err != nil {
		return nil, err
	}

	return s.envRepo.Create(ctx, env)
}

func (s *environmentServiceImpl) ListEnvironments(ctx context.Context, appID uuid.UUID) ([]*entity.Environment, error) {
	return s.envRepo.ListByApp(ctx, appID)
}

func (s *environmentServiceImpl) GetEnvironment(ctx context.Context, appID uuid.UUID, code string) (*entity.Environment, error) {
	env, err := s.envRepo.GetByAppAndCode(ctx, appID, code)
	if err != nil {
		return nil, err
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
