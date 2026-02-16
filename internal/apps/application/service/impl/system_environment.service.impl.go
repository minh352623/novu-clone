package impl

import (
	"context"

	"CONVERDA/internal/apps/application/service"
	"CONVERDA/internal/apps/domain/model/entity"
	"CONVERDA/internal/apps/domain/repository"
)

type systemEnvironmentServiceImpl struct {
	repo repository.SystemEnvironmentRepository
}

func NewSystemEnvironmentService(repo repository.SystemEnvironmentRepository) service.SystemEnvironmentService {
	return &systemEnvironmentServiceImpl{repo: repo}
}

func (s *systemEnvironmentServiceImpl) ListAll(ctx context.Context) ([]*entity.SystemEnvironment, error) {
	return s.repo.ListAll(ctx)
}
