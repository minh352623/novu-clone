package model

import (
	"time"

	"github.com/google/uuid"
)

type NotificationGroupModel struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:generate_uuid_v7()"`
	EnvironmentID uuid.UUID `gorm:"type:uuid;not null"`
	Name          string    `gorm:"type:text;not null"`
	Key           string    `gorm:"type:text;not null"`
	Description   string    `gorm:"type:text"`
	IsDefault     bool      `gorm:"default:false"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

func (NotificationGroupModel) TableName() string {
	return "notification_groups"
}
