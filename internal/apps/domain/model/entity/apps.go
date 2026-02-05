package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAppNotFound          = errors.New("app not found")
	ErrEnvironmentNotFound  = errors.New("environment not found")
	ErrDuplicateEnvironment = errors.New("environment already exists for this app")
)

type App struct {
	ID           uuid.UUID      `json:"id"`
	TenantID     uuid.UUID      `json:"tenant_id"`
	Name         string         `json:"name"`
	Description  *string        `json:"description,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	Environments []*Environment `json:"environments,omitempty"`
}

func NewApp(tenantID uuid.UUID, name string, description *string) (*App, error) {
	if name == "" {
		return nil, errors.New("app name is required")
	}
	return &App{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

type Environment struct {
	ID              uuid.UUID `json:"id"`
	AppID           uuid.UUID `json:"app_id"`
	EnvironmentCode string    `json:"environment_code"`
	APIKey          string    `json:"api_key"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func NewEnvironment(appID uuid.UUID, code string) (*Environment, error) {
	if code == "" {
		return nil, errors.New("environment code is required")
	}
	// TODO: Generate secure API Key
	apiKey := "sk_" + uuid.New().String()

	return &Environment{
		ID:              uuid.New(),
		AppID:           appID,
		EnvironmentCode: code,
		APIKey:          apiKey,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}, nil
}
