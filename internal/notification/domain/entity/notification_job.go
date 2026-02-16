package entity

import (
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusScheduled  JobStatus = "scheduled"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusCancelled  JobStatus = "cancelled"
)

type NotificationJob struct {
	ID             uuid.UUID
	EnvironmentID  uuid.UUID
	TenantID       *uuid.UUID
	Channel        string
	TemplateID     *uuid.UUID
	Status         JobStatus
	TotalCount     int
	SuccessCount   int
	FailedCount    int
	ScheduledAt    *time.Time
	StartedAt      *time.Time
	CompletedAt    *time.Time
	ErrorMessage   *string
	Metadata       map[string]interface{}
	RecipientsData map[string]interface{} // Flexible JSON structure
	CreatedBy      *uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
