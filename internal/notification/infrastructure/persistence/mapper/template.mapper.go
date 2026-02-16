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
