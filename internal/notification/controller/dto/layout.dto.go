package dto

import (
	"time"

	"CONVERDA/internal/notification/domain/entity"

	"github.com/google/uuid"
)

type CreateLayoutRequest struct {
	Name            string                 `json:"name" binding:"required"`
	Description     string                 `json:"description"`
	ContentHTML     string                 `json:"content_html" binding:"required"`
	VariablesSchema map[string]interface{} `json:"variables_schema"`
	IsDefault       bool                   `json:"is_default"`
}

type UpdateLayoutRequest struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	ContentHTML     string                 `json:"content_html"`
	VariablesSchema map[string]interface{} `json:"variables_schema"`
	IsDefault       bool                   `json:"is_default"`
}

type LayoutResponse struct {
	ID              uuid.UUID              `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	ContentHTML     string                 `json:"content_html"`
	VariablesSchema map[string]interface{} `json:"variables_schema"`
	IsDefault       bool                   `json:"is_default"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

func ToLayoutResponse(l *entity.NotificationLayout) LayoutResponse {
	if l == nil {
		return LayoutResponse{}
	}
	return LayoutResponse{
		ID:              l.ID,
		Name:            l.Name,
		Description:     l.Description,
		ContentHTML:     l.ContentHTML,
		VariablesSchema: l.VariablesSchema,
		IsDefault:       l.IsDefault,
		CreatedAt:       l.CreatedAt,
		UpdatedAt:       l.UpdatedAt,
	}
}

func ToLayoutResponseList(layouts []*entity.NotificationLayout) []LayoutResponse {
	var res []LayoutResponse
	for _, l := range layouts {
		res = append(res, ToLayoutResponse(l))
	}
	return res
}
