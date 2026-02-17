package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type WebhookModel struct {
	ID                  uuid.UUID      `gorm:"column:id;type:uuid;primaryKey"`
	TenantID            uuid.UUID      `gorm:"column:tenant_id;type:uuid;not null"`
	AppID               uuid.UUID      `gorm:"column:app_id;type:uuid;not null"`
	EnvironmentID       uuid.UUID      `gorm:"column:environment_id;type:uuid;not null"`
	URL                 string         `gorm:"column:url;type:text;not null"`
	Secret              string         `gorm:"column:secret;type:text;not null"`
	Description         *string        `gorm:"column:description;type:text"`
	Events              datatypes.JSON `gorm:"column:events;type:jsonb;default:'[]'"`
	IsActive            bool           `gorm:"column:is_active;default:true"`
	MaxRetries          int            `gorm:"column:max_retries;default:3"`
	RetryBackoffSeconds int            `gorm:"column:retry_backoff_seconds;default:5"`
	RetryTimeoutSeconds int            `gorm:"column:retry_timeout_seconds;default:10"`
	CreatedAt           time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt           time.Time      `gorm:"column:updated_at;autoUpdateTime"`

	// Relations
	App         *AppModel         `gorm:"foreignKey:AppID"`
	Environment *EnvironmentModel `gorm:"foreignKey:EnvironmentID"`
}

func (WebhookModel) TableName() string {
	return "webhooks"
}
