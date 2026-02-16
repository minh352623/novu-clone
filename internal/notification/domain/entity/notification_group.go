package entity

import (
	"time"

	"github.com/google/uuid"
)

type NotificationGroup struct {
	ID            uuid.UUID
	EnvironmentID uuid.UUID
	Name          string
	Key           string
	Description   string
	IsDefault     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
