package dto

import (
	"time"

	"CONVERDA/internal/apps/domain/model/entity"

	"github.com/google/uuid"
)

type CreateAppRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

type UpdateAppRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type CreateEnvironmentRequest struct {
	Code string `json:"code" binding:"required"`
}

type APIKeyResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Prefix    string     `json:"prefix"`
	Suffix    string     `json:"suffix"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type APIKeyFullResponse struct {
	APIKeyResponse
	PlainKey string `json:"plain_key,omitempty"` // Only returned once on creation/rotation
}

type EnvironmentResponse struct {
	ID              uuid.UUID         `json:"id"`
	AppID           uuid.UUID         `json:"app_id"`
	EnvironmentCode string            `json:"environment_code"`
	Keys            []*APIKeyResponse `json:"keys,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
}

type RotateKeyRequest struct {
	Name string `json:"name" binding:"required"`
}

type AppResponse struct {
	ID           uuid.UUID              `json:"id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	Name         string                 `json:"name"`
	Description  *string                `json:"description,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Environments []*EnvironmentResponse `json:"environments,omitempty"`
}

func ToEnvironmentResponse(env *entity.Environment) *EnvironmentResponse {
	if env == nil {
		return nil
	}
	resp := &EnvironmentResponse{
		ID:              env.ID,
		AppID:           env.AppID,
		EnvironmentCode: env.EnvironmentCode,
		CreatedAt:       env.CreatedAt,
	}

	if len(env.Keys) > 0 {
		resp.Keys = make([]*APIKeyResponse, 0, len(env.Keys))
		for _, k := range env.Keys {
			resp.Keys = append(resp.Keys, ToAPIKeyResponse(k))
		}
	}

	return resp
}

func ToEnvironmentResponseList(envs []*entity.Environment) []*EnvironmentResponse {
	list := make([]*EnvironmentResponse, 0, len(envs))
	for _, e := range envs {
		list = append(list, ToEnvironmentResponse(e))
	}
	return list
}

func ToAPIKeyResponse(k *entity.APIKey) *APIKeyResponse {
	if k == nil {
		return nil
	}
	return &APIKeyResponse{
		ID:        k.ID,
		Name:      k.Name,
		Prefix:    k.KeyPrefix,
		Suffix:    k.KeySuffix,
		ExpiresAt: k.ExpiresAt,
		RevokedAt: k.RevokedAt,
		CreatedAt: k.CreatedAt,
	}
}

func ToAPIKeyResponseList(keys []*entity.APIKey) []*APIKeyResponse {
	list := make([]*APIKeyResponse, 0, len(keys))
	for _, k := range keys {
		list = append(list, ToAPIKeyResponse(k))
	}
	return list
}

func ToAppResponse(app *entity.App) *AppResponse {
	if app == nil {
		return nil
	}
	resp := &AppResponse{
		ID:          app.ID,
		TenantID:    app.TenantID,
		Name:        app.Name,
		Description: app.Description,
		CreatedAt:   app.CreatedAt,
		UpdatedAt:   app.UpdatedAt,
	}
	if len(app.Environments) > 0 {
		resp.Environments = ToEnvironmentResponseList(app.Environments)
	}
	return resp
}

func ToAppResponseList(apps []*entity.App) []*AppResponse {
	list := make([]*AppResponse, 0, len(apps))
	for _, a := range apps {
		list = append(list, ToAppResponse(a))
	}
	return list
}
