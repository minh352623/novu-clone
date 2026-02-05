package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	ProviderTypeEmail = "email"
	ProviderTypeSMS   = "sms"
	ProviderTypePush  = "push"
	ProviderTypeInApp = "in-app"
)

type Provider struct {
	ID            uuid.UUID              `json:"id"`
	TenantID      uuid.UUID              `json:"tenant_id"`
	AppID         uuid.UUID              `json:"app_id"`
	EnvironmentID uuid.UUID              `json:"environment_id"`
	ProviderType  string                 `json:"provider_type"`
	ProviderName  string                 `json:"provider_name"`
	IsActive      bool                   `json:"is_active"`
	Configuration map[string]interface{} `json:"configuration"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

func NewProvider(tenantID, appID, environmentID uuid.UUID, providerType, providerName string, config map[string]interface{}) (*Provider, error) {
	if providerType == "" || providerName == "" {
		return nil, errors.New("provider type and name are required")
	}

	// Validate provider type
	switch providerType {
	case ProviderTypeEmail, ProviderTypeSMS, ProviderTypePush, ProviderTypeInApp:
		// Valid
	default:
		return nil, errors.New("invalid provider type: must be email, sms, push, or in-app")
	}

	return &Provider{
		ID:            uuid.New(),
		TenantID:      tenantID,
		AppID:         appID,
		EnvironmentID: environmentID,
		ProviderType:  providerType,
		ProviderName:  providerName,
		IsActive:      true,
		Configuration: config,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
}
