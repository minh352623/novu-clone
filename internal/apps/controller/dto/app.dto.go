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

type EnvironmentResponse struct {
	ID              uuid.UUID `json:"id"`
	AppID           uuid.UUID `json:"app_id"`
	EnvironmentCode string    `json:"environment_code"`
	APIKey          string    `json:"api_key"` // Show only once optionally? For now showing it.
	CreatedAt       time.Time `json:"created_at"`
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
	return &EnvironmentResponse{
		ID:              env.ID,
		AppID:           env.AppID,
		EnvironmentCode: env.EnvironmentCode,
		APIKey:          env.APIKey,
		CreatedAt:       env.CreatedAt,
	}
}

func ToEnvironmentResponseList(envs []*entity.Environment) []*EnvironmentResponse {
	list := make([]*EnvironmentResponse, 0, len(envs))
	for _, e := range envs {
		list = append(list, ToEnvironmentResponse(e))
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
