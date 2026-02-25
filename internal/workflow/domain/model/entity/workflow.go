package entity

import (
	"time"

	"github.com/google/uuid"
)

// Step types
const (
	StepTypeChannel = "channel"
	StepTypeDelay   = "delay"
	StepTypeDigest  = "digest"
)

// Workflow statuses
const (
	WorkflowExecutionRunning   = "running"
	WorkflowExecutionCompleted = "completed"
	WorkflowExecutionFailed    = "failed"
	WorkflowExecutionCancelled = "cancelled"
)

// Step execution statuses
const (
	StepStatusPending   = "pending"
	StepStatusRunning   = "running"
	StepStatusCompleted = "completed"
	StepStatusFailed    = "failed"
	StepStatusScheduled = "scheduled"
	StepStatusDigesting = "digesting"
)

// Workflow represents an automation workflow scoped to an environment.
type Workflow struct {
	ID                uuid.UUID
	TenantID          uuid.UUID
	EnvironmentID     uuid.UUID
	Name              string
	TriggerIdentifier string
	IsActive          bool
	Steps             []WorkflowStep
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// WorkflowStep represents a single step within a workflow.
type WorkflowStep struct {
	ID           uuid.UUID
	WorkflowID   uuid.UUID
	ParentStepID *uuid.UUID
	StepType     string // "channel", "delay", "digest"
	Config       map[string]interface{}
	Order        int
	CreatedAt    time.Time
}

// WorkflowExecution tracks one run of a workflow for a subscriber.
type WorkflowExecution struct {
	ID             uuid.UUID
	WorkflowID     uuid.UUID
	SubscriberKey  string
	TriggerPayload map[string]interface{}
	Status         string // running, completed, failed, cancelled
	CurrentStepID  *uuid.UUID
	StartedAt      time.Time
	CompletedAt    *time.Time
}

// StepExecution tracks execution of an individual step within a workflow run.
type StepExecution struct {
	ID          uuid.UUID
	ExecutionID uuid.UUID
	StepID      uuid.UUID
	Status      string // pending, running, completed, failed, scheduled, digesting
	ScheduledAt *time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	Output      map[string]interface{}
}

// DigestEvent represents a single buffered event during a digest window.
type DigestEvent struct {
	ID            uuid.UUID
	StepID        uuid.UUID
	ExecutionID   uuid.UUID
	SubscriberKey string
	Payload       map[string]interface{}
	CreatedAt     time.Time
}
