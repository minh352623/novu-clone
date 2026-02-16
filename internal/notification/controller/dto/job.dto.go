package dto

import (
	"time"

	"github.com/google/uuid"
)

type ScheduleJobRequest struct {
	EnvironmentID string                   `json:"environment_id"` // Optional in body if in header
	Channel       string                   `json:"channel" binding:"required"`
	TemplateID    string                   `json:"template_id"`
	ScheduledAt   *time.Time               `json:"scheduled_at"`
	Recipients    []map[string]interface{} `json:"recipients" binding:"required"`
	Metadata      map[string]interface{}   `json:"metadata"`
}

type JobResponse struct {
	ID             uuid.UUID              `json:"id"`
	EnvironmentID  uuid.UUID              `json:"environment_id"`
	Channel        string                 `json:"channel"`
	TemplateID     *uuid.UUID             `json:"template_id,omitempty"`
	Status         string                 `json:"status"`
	TotalCount     int                    `json:"total_count"`
	SuccessCount   int                    `json:"success_count"`
	FailedCount    int                    `json:"failed_count"`
	ScheduledAt    *time.Time             `json:"scheduled_at,omitempty"`
	StartedAt      *time.Time             `json:"started_at,omitempty"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
	ErrorMessage   *string                `json:"error_message,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	RecipientsData map[string]interface{} `json:"recipients_data,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

type ListJobsResponse struct {
	Jobs  []*JobResponse `json:"jobs"`
	Total int64          `json:"total"`
}
