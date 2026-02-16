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
	Keys            []*APIKey `json:"keys,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func NewEnvironment(appID uuid.UUID, code string) (*Environment, error) {
	if code == "" {
		return nil, errors.New("environment code is required")
	}
	// Initial key generation
	return &Environment{
		ID:              uuid.New(),
		AppID:           appID,
		EnvironmentCode: code,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}, nil
}

type APIKey struct {
	ID            uuid.UUID  `json:"id"`
	AppID         uuid.UUID  `json:"app_id"`
	EnvironmentID uuid.UUID  `json:"environment_id"`
	KeyHash       string     `json:"-"` // Never expose hash
	KeyPrefix     string     `json:"key_prefix"`
	KeySuffix     string     `json:"key_suffix"`
	Name          string     `json:"name"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func NewAPIKey(appID, envID uuid.UUID, name, prefix, suffix, hash string, ttl *time.Duration) *APIKey {
	var expiresAt *time.Time
	if ttl != nil {
		t := time.Now().Add(*ttl)
		expiresAt = &t
	}

	return &APIKey{
		ID:            uuid.New(),
		AppID:         appID,
		EnvironmentID: envID,
		KeyHash:       hash,
		KeyPrefix:     prefix,
		KeySuffix:     suffix,
		Name:          name,
		ExpiresAt:     expiresAt,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func (k *APIKey) IsActive() bool {
	if k.RevokedAt != nil {
		return false
	}
	if k.ExpiresAt != nil && time.Now().After(*k.ExpiresAt) {
		return false
	}
	return true
}

func (k *APIKey) Revoke() {
	now := time.Now()
	k.RevokedAt = &now
	k.UpdatedAt = now
}
