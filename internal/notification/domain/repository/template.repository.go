package repository

import (
	"context"

	"CONVERDA/internal/notification/domain/entity"

	"github.com/google/uuid"
)

type TemplateRepository interface {
	GetByCode(ctx context.Context, envID uuid.UUID, code, lang string) (*entity.Template, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Template, error)
	CreateContent(ctx context.Context, content *entity.TemplateContent) error
	UpdateContent(ctx context.Context, content *entity.TemplateContent) error
	GetContent(ctx context.Context, templateID uuid.UUID, version int, lang string) (*entity.TemplateContent, error)
	ListLanguages(ctx context.Context, templateID uuid.UUID, version int) ([]string, error)
}
