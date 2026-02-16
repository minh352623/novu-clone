package entity

import (
	"time"

	"github.com/google/uuid"
)

type NotificationLayout struct {
	ID              uuid.UUID
	EnvironmentID   uuid.UUID
	Name            string
	Description     string
	ContentHTML     string
	VariablesSchema map[string]interface{}
	IsDefault       bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
