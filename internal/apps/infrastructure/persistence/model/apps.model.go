package model

import (
	"time"

	"github.com/google/uuid"
)

type AppModel struct {
	ID                  uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	TenantID            uuid.UUID `gorm:"column:tenant_id;type:uuid;not null"`
	Name                string    `gorm:"column:name;type:text;not null"`
	Description         *string   `gorm:"column:description;type:text"`
	CreatedAt           time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt           time.Time `gorm:"column:updated_at;autoUpdateTime"`
	SLAThresholdSeconds int       `gorm:"column:sla_threshold_seconds;type:integer;default:0"`

	// Relations
	Environments []EnvironmentModel `gorm:"foreignKey:AppID"`
}

func (AppModel) TableName() string {
	return "apps"
}

type EnvironmentModel struct {
	ID                  uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	AppID               uuid.UUID `gorm:"column:app_id;type:uuid;not null"`
	EnvironmentCode     string    `gorm:"column:environment_code;type:text;not null"`
	CreatedAt           time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt           time.Time `gorm:"column:updated_at;autoUpdateTime"`
	SLAThresholdSeconds int       `gorm:"column:sla_threshold_seconds;type:integer;default:0"`
	RateLimitRPM        int       `gorm:"column:rate_limit_rpm;type:integer;default:0"`
	RateLimitDaily      int       `gorm:"column:rate_limit_daily;type:integer;default:0"`

	// Relation
	App  *AppModel     `gorm:"foreignKey:AppID"`
	Keys []APIKeyModel `gorm:"foreignKey:EnvironmentID"`
}

func (EnvironmentModel) TableName() string {
	return "environments"
}

type APIKeyModel struct {
	ID            uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	AppID         uuid.UUID  `gorm:"column:app_id;type:uuid;not null"`
	EnvironmentID uuid.UUID  `gorm:"column:environment_id;type:uuid;not null"`
	KeyHash       string     `gorm:"column:key_hash;type:text;unique;not null"`
	KeyPrefix     string     `gorm:"column:key_prefix;type:text;not null"`
	KeySuffix     string     `gorm:"column:key_suffix;type:text;not null"`
	Name          string     `gorm:"column:name;type:text;not null"`
	ExpiresAt     *time.Time `gorm:"column:expires_at"`
	RevokedAt     *time.Time `gorm:"column:revoked_at"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime"`

	// Relation
	Environment *EnvironmentModel `gorm:"foreignKey:EnvironmentID"`
}

func (APIKeyModel) TableName() string {
	return "app_api_keys"
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
