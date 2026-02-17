package repository

import (
	"context"
	"errors"
	"fmt"

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

	// Fetch Content for requested language
	var c model.NotificationTemplateContentModel
	if err := r.db.WithContext(ctx).Where("template_id = ? AND version = ? AND language_code = ?", t.ID, t.ActiveVersion, language).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Fallback to default language "en"
			if language != "en" {
				if err2 := r.db.WithContext(ctx).Where("template_id = ? AND version = ? AND language_code = ?", t.ID, t.ActiveVersion, "en").First(&c).Error; err2 != nil {
					if errors.Is(err2, gorm.ErrRecordNotFound) {
						return mapper.ToTemplateDomain(&t, nil), nil
					}
					return nil, err2
				}
				return mapper.ToTemplateDomain(&t, &c), nil
			}
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

func (r *templateRepository) CreateContent(ctx context.Context, content *entity.TemplateContent) error {
	// Check for duplicate (same template + version + language)
	var count int64
	r.db.WithContext(ctx).Model(&model.NotificationTemplateContentModel{}).
		Where("template_id = ? AND version = ? AND language_code = ?", content.TemplateID, content.Version, content.LanguageCode).
		Count(&count)
	if count > 0 {
		return fmt.Errorf("content for language '%s' already exists for this template version", content.LanguageCode)
	}

	m := mapper.ToTemplateContentModel(content)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *templateRepository) UpdateContent(ctx context.Context, content *entity.TemplateContent) error {
	m := mapper.ToTemplateContentModel(content)
	return r.db.WithContext(ctx).
		Where("id = ?", m.ID).
		Updates(map[string]interface{}{
			"subject":   m.Subject,
			"body_text": m.BodyText,
			"body_html": m.BodyHtml,
			"body_push": m.BodyPush,
		}).Error
}

func (r *templateRepository) GetContent(ctx context.Context, templateID uuid.UUID, version int, lang string) (*entity.TemplateContent, error) {
	var c model.NotificationTemplateContentModel
	if err := r.db.WithContext(ctx).
		Where("template_id = ? AND version = ? AND language_code = ?", templateID, version, lang).
		First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToTemplateContentDomain(&c), nil
}

func (r *templateRepository) ListLanguages(ctx context.Context, templateID uuid.UUID, version int) ([]string, error) {
	var langs []string
	if err := r.db.WithContext(ctx).
		Model(&model.NotificationTemplateContentModel{}).
		Where("template_id = ? AND version = ?", templateID, version).
		Distinct("language_code").
		Pluck("language_code", &langs).Error; err != nil {
		return nil, err
	}
	return langs, nil
}
