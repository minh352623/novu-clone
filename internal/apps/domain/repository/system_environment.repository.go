package repository

import (
	"context"

	"CONVERDA/internal/apps/domain/model/entity"
)

type SystemEnvironmentRepository interface {
	ListAll(ctx context.Context) ([]*entity.SystemEnvironment, error)
	GetByCode(ctx context.Context, code string) (*entity.SystemEnvironment, error)
}
