package dto

import (
	"time"

	"CONVERDA/internal/workflow/domain/model/entity"

	"github.com/google/uuid"
)

// --- Request DTOs ---

type CreateWorkflowRequest struct {
	Name              string `json:"name" binding:"required"`
	TriggerIdentifier string `json:"trigger_identifier" binding:"required"`
}

type UpdateWorkflowRequest struct {
	Name              *string `json:"name"`
	TriggerIdentifier *string `json:"trigger_identifier"`
}

type CreateStepRequest struct {
	StepType     string                 `json:"step_type" binding:"required,oneof=channel delay digest"`
	Config       map[string]interface{} `json:"config"`
	Order        int                    `json:"order"`
	ParentStepID *uuid.UUID             `json:"parent_step_id,omitempty"`
}

type UpdateStepRequest struct {
	StepType *string                `json:"step_type"`
	Config   map[string]interface{} `json:"config"`
	Order    *int                   `json:"order"`
}

type TriggerWorkflowRequest struct {
	EnvironmentID     uuid.UUID              `json:"environment_id" binding:"required"`
	TriggerIdentifier string                 `json:"trigger_identifier" binding:"required"`
	SubscriberKey     string                 `json:"subscriber_key" binding:"required"`
	Payload           map[string]interface{} `json:"payload"`
}

// --- Response DTOs ---

type WorkflowResponse struct {
	ID                uuid.UUID      `json:"id"`
	EnvironmentID     uuid.UUID      `json:"environment_id"`
	Name              string         `json:"name"`
	TriggerIdentifier string         `json:"trigger_identifier"`
	IsActive          bool           `json:"is_active"`
	Steps             []StepResponse `json:"steps"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

type StepResponse struct {
	ID           uuid.UUID              `json:"id"`
	WorkflowID   uuid.UUID              `json:"workflow_id"`
	ParentStepID *uuid.UUID             `json:"parent_step_id,omitempty"`
	StepType     string                 `json:"step_type"`
	Config       map[string]interface{} `json:"config"`
	Order        int                    `json:"order"`
	CreatedAt    time.Time              `json:"created_at"`
}

// --- Mapping functions ---

func ToWorkflowResponse(w *entity.Workflow) *WorkflowResponse {
	if w == nil {
		return nil
	}
	steps := make([]StepResponse, 0, len(w.Steps))
	for _, s := range w.Steps {
		steps = append(steps, ToStepResponse(&s))
	}
	return &WorkflowResponse{
		ID:                w.ID,
		EnvironmentID:     w.EnvironmentID,
		Name:              w.Name,
		TriggerIdentifier: w.TriggerIdentifier,
		IsActive:          w.IsActive,
		Steps:             steps,
		CreatedAt:         w.CreatedAt,
		UpdatedAt:         w.UpdatedAt,
	}
}

func ToStepResponse(s *entity.WorkflowStep) StepResponse {
	return StepResponse{
		ID:           s.ID,
		WorkflowID:   s.WorkflowID,
		ParentStepID: s.ParentStepID,
		StepType:     s.StepType,
		Config:       s.Config,
		Order:        s.Order,
		CreatedAt:    s.CreatedAt,
	}
}

func ToWorkflowResponseList(workflows []*entity.Workflow) []*WorkflowResponse {
	list := make([]*WorkflowResponse, 0, len(workflows))
	for _, w := range workflows {
		list = append(list, ToWorkflowResponse(w))
	}
	return list
}
