package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID            uuid.UUID       `json:"id"`
	TenantID      uuid.UUID       `json:"tenant_id"`
	EnvironmentID uuid.UUID       `json:"environment_id"`
	TemplateCode  string          `json:"template_code"`
	Recipient     string          `json:"recipient"` // Email, Phone, Token
	Channel       string          `json:"channel"`   // email, sms, push
	Status        string          `json:"status"`    // pending, sent, failed
	Data          json.RawMessage `json:"data"`      // Variables
	ErrorMessage  *string         `json:"error_message,omitempty"`
	SentAt        *time.Time      `json:"sent_at,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type Template struct {
	ID        uuid.UUID  `json:"id"`
	Code      string     `json:"code"`
	GroupID   *uuid.UUID `json:"group_id"`
	LayoutID  *uuid.UUID `json:"layout_id"`
	Version   int        `json:"version"`
	Subject   string     `json:"subject"` // For email/push title
	Body      string     `json:"body"`    // HTML or Text
	Language  string     `json:"language"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// TemplateContent represents a single language version of template content.
type TemplateContent struct {
	ID           uuid.UUID       `json:"id"`
	TemplateID   uuid.UUID       `json:"template_id"`
	Version      int             `json:"version"`
	LanguageCode string          `json:"language_code"`
	Subject      string          `json:"subject"`
	BodyText     string          `json:"body_text"`
	BodyHTML     string          `json:"body_html"`
	BodyPush     json.RawMessage `json:"body_push,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

func NewTemplateContent(templateID uuid.UUID, version int, langCode, subject, bodyText, bodyHTML string, bodyPush json.RawMessage) *TemplateContent {
	return &TemplateContent{
		ID:           uuid.New(),
		TemplateID:   templateID,
		Version:      version,
		LanguageCode: langCode,
		Subject:      subject,
		BodyText:     bodyText,
		BodyHTML:     bodyHTML,
		BodyPush:     bodyPush,
		CreatedAt:    time.Now(),
	}
}

type ProviderConfig struct {
	ID            uuid.UUID       `json:"id"`
	EnvironmentID uuid.UUID       `json:"environment_id"`
	Type          string          `json:"type"` // smtp, fcm, twilio
	Config        json.RawMessage `json:"config"`
	IsActive      bool            `json:"is_active"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

func NewNotification(tenantID, envID uuid.UUID, code, recipient, channel string, data []byte) *Notification {
	return &Notification{
		ID:            uuid.New(),
		TenantID:      tenantID,
		EnvironmentID: envID,
		TemplateCode:  code,
		Recipient:     recipient,
		Channel:       channel,
		Status:        "pending",
		Data:          data,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}
