package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// WorkflowModel maps to the `workflows` table.
type WorkflowModel struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:generate_uuid_v7()"`
	EnvironmentID     uuid.UUID `gorm:"type:uuid;not null"`
	Name              string    `gorm:"type:text;not null"`
	TriggerIdentifier string    `gorm:"column:trigger_identifier;type:text;not null"`
	IsActive          bool      `gorm:"default:false"`
	CreatedAt         time.Time
	UpdatedAt         time.Time

	// Relations
	Steps []WorkflowStepModel `gorm:"foreignKey:WorkflowID"`
}

func (WorkflowModel) TableName() string {
	return "workflows"
}

// WorkflowStepModel maps to the `workflow_steps` table.
type WorkflowStepModel struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:generate_uuid_v7()"`
	WorkflowID   uuid.UUID      `gorm:"type:uuid;not null"`
	ParentStepID *uuid.UUID     `gorm:"type:uuid"`
	StepType     string         `gorm:"type:text;not null"`
	Config       datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
	Order        int            `gorm:"column:order;not null;default:0"`
	CreatedAt    time.Time
}

func (WorkflowStepModel) TableName() string {
	return "workflow_steps"
}
