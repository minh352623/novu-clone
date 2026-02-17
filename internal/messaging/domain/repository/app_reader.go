package repository

import (
	"context"

	"github.com/google/uuid"
)

// AppInfo is a DTO for application data
type AppInfo struct {
	ID                  uuid.UUID
	TenantID            uuid.UUID
	Name                string
	SLAThresholdSeconds int
}

// EnvInfo is a DTO for environment data
type EnvInfo struct {
	ID                  uuid.UUID
	AppID               uuid.UUID
	Code                string
	SLAThresholdSeconds int
}

// AppReader defines the interface for reading application and environment data from other modules
type AppReader interface {
	GetApp(ctx context.Context, id uuid.UUID) (*AppInfo, error)
	GetEnvironment(ctx context.Context, id uuid.UUID) (*EnvInfo, error)
}
