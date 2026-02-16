package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// NotificationTemplateModel maps to 'notification_templates'
type NotificationTemplateModel struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:generate_uuid_v7()"`
	EnvironmentID uuid.UUID  `gorm:"type:uuid;not null"`
	TenantID      *uuid.UUID `gorm:"type:uuid"`
	TemplateCode  string     `gorm:"not null"`
	GroupID       *uuid.UUID `gorm:"type:uuid"`
	LayoutID      *uuid.UUID `gorm:"type:uuid"`
	Channel       string     `gorm:"not null"`
	ActiveVersion int        `gorm:"default:1"`
	IsDeleted     bool       `gorm:"default:false"`
	CreatedAt     time.Time
	UpdatedAt     time.Time

	// Relations
	Contents []NotificationTemplateContentModel `gorm:"foreignKey:TemplateID"`
}

func (NotificationTemplateModel) TableName() string {
	return "notification_templates"
}

// NotificationTemplateContentModel maps to 'notification_template_contents'
type NotificationTemplateContentModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:generate_uuid_v7()"`
	TemplateID   uuid.UUID `gorm:"type:uuid;not null"`
	Version      int       `gorm:"not null"`
	LanguageCode string    `gorm:"not null;default:'en'"`
	Subject      string
	BodyText     string
	BodyHtml     string
	BodyPush     json.RawMessage `gorm:"type:jsonb"`
	CreatedAt    time.Time
}

func (NotificationTemplateContentModel) TableName() string {
	return "notification_template_contents"
}
