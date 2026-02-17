package mapper

import (
	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/infrastructure/persistence/model"
)

// Template Logic is complex because we join Template + Content.
// We will handle logic in Repository to select right content, and map here.

func ToTemplateDomain(t *model.NotificationTemplateModel, c *model.NotificationTemplateContentModel) *entity.Template {
	if t == nil {
		return nil
	}

	var subject, body, lang string
	if c != nil {
		subject = c.Subject
		body = c.BodyHtml
		if body == "" {
			body = c.BodyText
		}
		lang = c.LanguageCode
	}

	return &entity.Template{
		ID:        t.ID,
		Code:      t.TemplateCode,
		GroupID:   t.GroupID,
		LayoutID:  t.LayoutID,
		Version:   t.ActiveVersion,
		Subject:   subject,
		Body:      body,
		Language:  lang,
		IsActive:  !t.IsDeleted,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

// ToTemplateContentDomain maps a content model to the domain entity.
func ToTemplateContentDomain(c *model.NotificationTemplateContentModel) *entity.TemplateContent {
	if c == nil {
		return nil
	}
	return &entity.TemplateContent{
		ID:           c.ID,
		TemplateID:   c.TemplateID,
		Version:      c.Version,
		LanguageCode: c.LanguageCode,
		Subject:      c.Subject,
		BodyText:     c.BodyText,
		BodyHTML:     c.BodyHtml,
		BodyPush:     c.BodyPush,
		CreatedAt:    c.CreatedAt,
	}
}

// ToTemplateContentModel maps a domain entity to the GORM model.
func ToTemplateContentModel(e *entity.TemplateContent) *model.NotificationTemplateContentModel {
	if e == nil {
		return nil
	}
	return &model.NotificationTemplateContentModel{
		ID:           e.ID,
		TemplateID:   e.TemplateID,
		Version:      e.Version,
		LanguageCode: e.LanguageCode,
		Subject:      e.Subject,
		BodyText:     e.BodyText,
		BodyHtml:     e.BodyHTML,
		BodyPush:     e.BodyPush,
		CreatedAt:    e.CreatedAt,
	}
}
