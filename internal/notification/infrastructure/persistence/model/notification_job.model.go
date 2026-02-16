package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type NotificationJobModel struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:generate_uuid_v7()"`
	EnvironmentID  uuid.UUID      `gorm:"type:uuid;not null"`
	TenantID       *uuid.UUID     `gorm:"type:uuid"`
	Channel        string         `gorm:"type:text;not null"`
	TemplateID     *uuid.UUID     `gorm:"type:uuid"`
	Status         string         `gorm:"type:text;not null;default:'pending'"`
	TotalCount     int            `gorm:"type:int;default:0"`
	SuccessCount   int            `gorm:"type:int;default:0"`
	FailedCount    int            `gorm:"type:int;default:0"`
	ScheduledAt    *time.Time     `gorm:"type:timestamptz"`
	StartedAt      *time.Time     `gorm:"type:timestamptz"`
	CompletedAt    *time.Time     `gorm:"type:timestamptz"`
	ErrorMessage   *string        `gorm:"type:text"`
	Metadata       datatypes.JSON `gorm:"type:jsonb"`
	RecipientsData datatypes.JSON `gorm:"type:jsonb"`
	CreatedBy      *uuid.UUID     `gorm:"type:uuid"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"`
}

func (NotificationJobModel) TableName() string {
	return "notification_jobs"
}
