package repository

import (
	"context"
	"errors"

	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/domain/repository"
	"CONVERDA/internal/notification/infrastructure/persistence/mapper"
	"CONVERDA/internal/notification/infrastructure/persistence/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type templateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) repository.TemplateRepository {
	return &templateRepository{db: db}
}

func (r *templateRepository) GetByCode(ctx context.Context, envID uuid.UUID, code string, language string) (*entity.Template, error) {
	var t model.NotificationTemplateModel
	if err := r.db.WithContext(ctx).Where("environment_id = ? AND template_code = ?", envID, code).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not found
		}
		return nil, err
	}

	// Fetch Content
	var c model.NotificationTemplateContentModel
	if err := r.db.WithContext(ctx).Where("template_id = ? AND version = ? AND language_code = ?", t.ID, t.ActiveVersion, language).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Content not found for this language? Fallback?
			// For now, return template with empty body or err?
			// Return entity with nil content (handled by mapper now)
			return mapper.ToTemplateDomain(&t, nil), nil
		}
		return nil, err
	}

	return mapper.ToTemplateDomain(&t, &c), nil
}

func (r *templateRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Template, error) {
	var t model.NotificationTemplateModel
	if err := r.db.WithContext(ctx).First(&t, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	// For GetByID, we only need the template metadata (Code, etc.) for JobScheduler
	// We don't necessarily need content.
	return mapper.ToTemplateDomain(&t, nil), nil
}
