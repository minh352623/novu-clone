package dto

import (
	"time"

	"CONVERDA/internal/notification/domain/entity"

	"github.com/google/uuid"
)

type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Key         string `json:"key" binding:"required,alphanum"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default"`
}

type UpdateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type GroupResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Key         string    `json:"key"`
	Description string    `json:"description"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToGroupResponse(g *entity.NotificationGroup) GroupResponse {
	if g == nil {
		return GroupResponse{}
	}
	return GroupResponse{
		ID:          g.ID,
		Name:        g.Name,
		Key:         g.Key,
		Description: g.Description,
		IsDefault:   g.IsDefault,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
}

func ToGroupResponseList(groups []*entity.NotificationGroup) []GroupResponse {
	var res []GroupResponse
	for _, g := range groups {
		res = append(res, ToGroupResponse(g))
	}
	return res
}
