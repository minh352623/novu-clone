package service

import (
	"context"

	"CONVERDA/internal/apps/domain/model/entity"
)

type SystemEnvironmentService interface {
	ListAll(ctx context.Context) ([]*entity.SystemEnvironment, error)
}
