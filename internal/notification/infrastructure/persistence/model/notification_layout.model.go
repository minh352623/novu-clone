package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type NotificationLayoutModel struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:generate_uuid_v7()"`
	EnvironmentID   uuid.UUID      `gorm:"type:uuid;not null"`
	Name            string         `gorm:"type:text;not null"`
	Description     string         `gorm:"type:text"`
	ContentHTML     string         `gorm:"type:text;not null"`
	VariablesSchema datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
	IsDefault       bool           `gorm:"default:false"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"`
}

func (NotificationLayoutModel) TableName() string {
	return "notification_layouts"
}
