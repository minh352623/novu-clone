package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type SubscriberModel struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:generate_uuid_v7()"`
	EnvironmentID uuid.UUID      `gorm:"type:uuid;not null"`
	SubscriberKey string         `gorm:"type:text;not null"`
	Email         *string        `gorm:"type:text"`
	Phone         *string        `gorm:"type:text"`
	Data          datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (SubscriberModel) TableName() string {
	return "subscribers"
}

type ThreadModel struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:generate_uuid_v7()"`
	EnvironmentID uuid.UUID      `gorm:"type:uuid;not null"`
	Type          string         `gorm:"type:text;not null;default:'support'"`
	Channel       string         `gorm:"type:text;not null;default:'web'"`
	Status        string         `gorm:"type:text;not null;default:'unassigned'"`
	Metadata      datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
	ReferenceHash *string        `gorm:"type:text;uniqueIndex"`
	IsOverdue     bool           `gorm:"type:boolean;not null;default:false"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (ThreadModel) TableName() string {
	return "threads"
}

type ThreadParticipantModel struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:generate_uuid_v7()"`
	ThreadID   uuid.UUID  `gorm:"type:uuid;not null"`
	EntityType string     `gorm:"type:text;not null"` // user, subscriber
	EntityID   uuid.UUID  `gorm:"type:uuid;not null"`
	LastReadAt *time.Time `gorm:"type:timestamp"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (ThreadParticipantModel) TableName() string {
	return "thread_participants"
}

type MessageModel struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:generate_uuid_v7()"`
	TenantID      uuid.UUID      `gorm:"type:uuid;not null"`
	EnvironmentID uuid.UUID      `gorm:"type:uuid;not null"`
	ThreadID      uuid.UUID      `gorm:"type:uuid;not null"`
	SenderType    string         `gorm:"type:text;not null"`
	SenderID      *uuid.UUID     `gorm:"type:uuid"`
	Content       datatypes.JSON `gorm:"type:jsonb;not null"`
	Type          string         `gorm:"type:text;not null;default:'standard'"`
	ParentID      *uuid.UUID     `gorm:"type:uuid"`
	CreatedAt     time.Time
}

func (MessageModel) TableName() string {
	return "messages"
}

type MessageClosureModel struct {
	AncestorID   uuid.UUID `gorm:"type:uuid;primary_key"`
	DescendantID uuid.UUID `gorm:"type:uuid;primary_key"`
	Depth        int       `gorm:"not null"`
}

func (MessageClosureModel) TableName() string {
	return "message_closure"
}

type AssignmentLogModel struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key;default:generate_uuid_v7()"`
	ThreadID            uuid.UUID  `gorm:"type:uuid;not null"`
	AssignedToMemberID  *uuid.UUID `gorm:"type:uuid"`
	AssignedAt          time.Time  `gorm:"default:now()"`
	ResolvedAt          *time.Time
	ResponseTimeSeconds *int
}

func (AssignmentLogModel) TableName() string {
	return "assignment_logs"
}
