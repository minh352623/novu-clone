package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ProviderConfigModel maps to 'provider_configs'
type ProviderConfigModel struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey;default:generate_uuid_v7()"`
	ProviderID    uuid.UUID       `gorm:"type:uuid;not null"`
	EnvironmentID uuid.UUID       `gorm:"type:uuid;not null"`
	TenantID      uuid.UUID       `gorm:"type:uuid;not null"`
	AppID         uuid.UUID       `gorm:"type:uuid;not null"`
	Configuration json.RawMessage `gorm:"type:jsonb;not null"`
	IsActive      bool            `gorm:"default:true"`
	CreatedAt     time.Time
	UpdatedAt     time.Time

	// Association
	Provider ProviderModel `gorm:"foreignKey:ProviderID"`
}

func (ProviderConfigModel) TableName() string {
	return "provider_configs"
}

type ProviderModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:generate_uuid_v7()"`
	ProviderType string    `gorm:"not null"`
	ProviderName string    `gorm:"not null"`
}

func (ProviderModel) TableName() string {
	return "providers"
}
