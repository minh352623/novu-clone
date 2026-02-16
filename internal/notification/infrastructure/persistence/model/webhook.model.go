package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type WebhookModel struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:generate_uuid_v7()"`
	TenantID      uuid.UUID      `gorm:"type:uuid;not null"`
	AppID         uuid.UUID      `gorm:"type:uuid;not null"`
	EnvironmentID uuid.UUID      `gorm:"type:uuid;not null"`
	URL           string         `gorm:"not null"`
	Secret        string         `gorm:"not null"`
	Description   *string        `gorm:"type:text"`
	Events        datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'"`
	IsActive      bool           `gorm:"default:true"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (WebhookModel) TableName() string {
	return "webhooks"
}

type WebhookLogModel struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:generate_uuid_v7()"`
	WebhookID      uuid.UUID      `gorm:"type:uuid;not null"`
	URL            string         `gorm:"not null"`
	EventType      string         `gorm:"not null"`
	RequestPayload datatypes.JSON `gorm:"type:jsonb"`
	ResponseCode   int
	ResponseBody   *string `gorm:"type:text"`
	DurationMs     int64
	Status         string    `gorm:"not null"`
	TenantID       uuid.UUID `gorm:"type:uuid;not null"`
	AppID          uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt      time.Time
}

func (WebhookLogModel) TableName() string {
	return "webhook_logs"
}
