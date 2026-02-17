package dto

import (
	"encoding/json"
	"time"

	"CONVERDA/internal/notification/domain/entity"

	"github.com/google/uuid"
)

// CreateTemplateContentRequest is used to add content in a new language.
type CreateTemplateContentRequest struct {
	LanguageCode string          `json:"language_code" binding:"required"`
	Subject      string          `json:"subject"`
	BodyText     string          `json:"body_text"`
	BodyHTML     string          `json:"body_html"`
	BodyPush     json.RawMessage `json:"body_push,omitempty"`
}

// UpdateTemplateContentRequest is used to update content for a language.
type UpdateTemplateContentRequest struct {
	Subject  *string          `json:"subject,omitempty"`
	BodyText *string          `json:"body_text,omitempty"`
	BodyHTML *string          `json:"body_html,omitempty"`
	BodyPush *json.RawMessage `json:"body_push,omitempty"`
}

// TemplateContentResponse is the API response for template content.
type TemplateContentResponse struct {
	ID           uuid.UUID `json:"id"`
	TemplateID   uuid.UUID `json:"template_id"`
	Version      int       `json:"version"`
	LanguageCode string    `json:"language_code"`
	Subject      string    `json:"subject"`
	BodyText     string    `json:"body_text"`
	BodyHTML     string    `json:"body_html"`
	CreatedAt    time.Time `json:"created_at"`
}

// AvailableLanguagesResponse lists the available languages for a template.
type AvailableLanguagesResponse struct {
	TemplateID uuid.UUID `json:"template_id"`
	Version    int       `json:"version"`
	Languages  []string  `json:"languages"`
}

// ToTemplateContentResponse maps a domain entity to the API response.
func ToTemplateContentResponse(c *entity.TemplateContent) *TemplateContentResponse {
	if c == nil {
		return nil
	}
	return &TemplateContentResponse{
		ID:           c.ID,
		TemplateID:   c.TemplateID,
		Version:      c.Version,
		LanguageCode: c.LanguageCode,
		Subject:      c.Subject,
		BodyText:     c.BodyText,
		BodyHTML:     c.BodyHTML,
		CreatedAt:    c.CreatedAt,
	}
}
