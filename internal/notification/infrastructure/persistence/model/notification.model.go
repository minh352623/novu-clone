package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type NotificationModel struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey;default:generate_uuid_v7()"`
	TenantID      uuid.UUID       `gorm:"type:uuid;not null"`
	EnvironmentID uuid.UUID       `gorm:"type:uuid;not null"`
	TemplateCode  string          `gorm:"not null"`
	Recipient     string          `gorm:"not null"`
	Channel       string          `gorm:"not null"`
	Status        string          `gorm:"not null;default:'pending'"`
	Data          json.RawMessage `gorm:"type:jsonb"`
	ErrorMessage  *string
	SentAt        *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (NotificationModel) TableName() string {
	return "notifications"
}
