package model

import (
	"time"

	"github.com/google/uuid"
)

type UsageMetricModel struct {
	ID            uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	AppID         uuid.UUID `gorm:"column:app_id;type:uuid;not null"`
	EnvironmentID uuid.UUID `gorm:"column:environment_id;type:uuid;not null"`
	ProviderType  string    `gorm:"column:provider_type;type:text;not null"`
	Direction     string    `gorm:"column:direction;type:text;not null;default:outbound"`
	StatusCode    int       `gorm:"column:status_code;type:integer"`
	Timestamp     time.Time `gorm:"column:timestamp;autoCreateTime"`
}

func (UsageMetricModel) TableName() string {
	return "usage_metrics"
}
