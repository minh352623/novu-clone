package repository

import (
	"context"

	"CONVERDA/internal/notification/domain/entity"

	"github.com/google/uuid"
)

type TemplateRepository interface {
	GetByCode(ctx context.Context, envID uuid.UUID, code, lang string) (*entity.Template, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Template, error)
}
