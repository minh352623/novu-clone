package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// WorkflowExecutionModel maps to the `workflow_executions` table.
type WorkflowExecutionModel struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:generate_uuid_v7()"`
	WorkflowID     uuid.UUID      `gorm:"type:uuid;not null"`
	SubscriberKey  string         `gorm:"type:text;not null"`
	TriggerPayload datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
	Status         string         `gorm:"type:text;not null;default:'running'"`
	CurrentStepID  *uuid.UUID     `gorm:"type:uuid"`
	StartedAt      time.Time
	CompletedAt    *time.Time
}

func (WorkflowExecutionModel) TableName() string {
	return "workflow_executions"
}

// StepExecutionModel maps to the `step_executions` table.
type StepExecutionModel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:generate_uuid_v7()"`
	ExecutionID uuid.UUID      `gorm:"type:uuid;not null"`
	StepID      uuid.UUID      `gorm:"type:uuid;not null"`
	Status      string         `gorm:"type:text;not null;default:'pending'"`
	ScheduledAt *time.Time     `gorm:"type:timestamptz"`
	StartedAt   *time.Time     `gorm:"type:timestamptz"`
	CompletedAt *time.Time     `gorm:"type:timestamptz"`
	Output      datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
}

func (StepExecutionModel) TableName() string {
	return "step_executions"
}

// DigestEventModel maps to the `digest_events` table.
type DigestEventModel struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:generate_uuid_v7()"`
	StepID        uuid.UUID      `gorm:"type:uuid;not null"`
	ExecutionID   uuid.UUID      `gorm:"type:uuid;not null"`
	SubscriberKey string         `gorm:"type:text;not null"`
	Payload       datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
	CreatedAt     time.Time
}

func (DigestEventModel) TableName() string {
	return "digest_events"
}
