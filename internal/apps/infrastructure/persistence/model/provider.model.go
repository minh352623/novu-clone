package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type ProviderModel struct {
	ID            uuid.UUID      `gorm:"column:id;type:uuid;primaryKey"`
	TenantID      uuid.UUID      `gorm:"column:tenant_id;type:uuid;not null"`
	AppID         uuid.UUID      `gorm:"column:app_id;type:uuid;not null"`
	EnvironmentID uuid.UUID      `gorm:"column:environment_id;type:uuid;not null"`
	ProviderType  string         `gorm:"column:provider_type;type:text;not null"`
	ProviderName  string         `gorm:"column:provider_name;type:text;not null"`
	IsActive      bool           `gorm:"column:is_active;default:true"`
	Configuration datatypes.JSON `gorm:"column:configuration;type:jsonb;not null"`
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;autoUpdateTime"`

	// Relations
	App         *AppModel         `gorm:"foreignKey:AppID"`
	Environment *EnvironmentModel `gorm:"foreignKey:EnvironmentID"`
}

func (ProviderModel) TableName() string {
	return "providers"
}
