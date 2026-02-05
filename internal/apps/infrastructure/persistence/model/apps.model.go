package model

import (
	"time"

	"github.com/google/uuid"
)

type AppModel struct {
	ID          uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	TenantID    uuid.UUID `gorm:"column:tenant_id;type:uuid;not null"`
	Name        string    `gorm:"column:name;type:text;not null"`
	Description *string   `gorm:"column:description;type:text"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`

	// Relations
	Environments []EnvironmentModel `gorm:"foreignKey:AppID"`
}

func (AppModel) TableName() string {
	return "apps"
}

type EnvironmentModel struct {
	ID              uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	AppID           uuid.UUID `gorm:"column:app_id;type:uuid;not null"`
	EnvironmentCode string    `gorm:"column:environment_code;type:text;not null"`
	APIKey          string    `gorm:"column:api_key;type:text;unique;not null"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime"`

	// Relation
	App *AppModel `gorm:"foreignKey:AppID"`
}

func (EnvironmentModel) TableName() string {
	return "environments"
}

type SystemEnvironmentModel struct {
	Code        string    `gorm:"column:code;primaryKey"`
	Name        string    `gorm:"column:name"`
	Description *string   `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (SystemEnvironmentModel) TableName() string {
	return "system_environments"
}
