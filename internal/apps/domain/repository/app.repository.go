package repository

import (
	"context"

	"CONVERDA/internal/apps/domain/model/entity"

	"github.com/google/uuid"
)

type AppRepository interface {
	Create(ctx context.Context, app *entity.App) (*entity.App, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.App, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.App, error)
	Update(ctx context.Context, app *entity.App) error
	Delete(ctx context.Context, id uuid.UUID) error
}
