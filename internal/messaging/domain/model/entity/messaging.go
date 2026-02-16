package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Subscriber struct {
	ID            uuid.UUID       `json:"id"`
	EnvironmentID uuid.UUID       `json:"environment_id"`
	SubscriberKey string          `json:"subscriber_key"`
	Email         *string         `json:"email"`
	Phone         *string         `json:"phone"`
	Data          json.RawMessage `json:"data"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type Thread struct {
	ID            uuid.UUID       `json:"id"`
	EnvironmentID uuid.UUID       `json:"environment_id"`
	Type          string          `json:"type"`    // direct, group, support
	Channel       string          `json:"channel"` // zalo, facebook, web
	Status        string          `json:"status"`  // unassigned, assigned, resolved
	Metadata      json.RawMessage `json:"metadata"`
	ReferenceHash *string         `json:"reference_hash,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`

	// Enriched
	Participants []*ThreadParticipant `json:"participants,omitempty"`
}

type ThreadParticipant struct {
	ID         uuid.UUID  `json:"id"`
	ThreadID   uuid.UUID  `json:"thread_id"`
	EntityType string     `json:"entity_type"` // user, subscriber
	EntityID   uuid.UUID  `json:"entity_id"`
	LastReadAt *time.Time `json:"last_read_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	// Enriched
	DisplayName string `json:"display_name,omitempty"`
}

type Message struct {
	ID            uuid.UUID       `json:"id"`
	TenantID      uuid.UUID       `json:"tenant_id"`
	EnvironmentID uuid.UUID       `json:"environment_id"`
	ThreadID      uuid.UUID       `json:"thread_id"`
	SenderType    string          `json:"sender_type"` // agent, contact, system
	SenderID      *uuid.UUID      `json:"sender_id"`
	Content       json.RawMessage `json:"content"`
	Type          string          `json:"type"` // standard, internal_note
	ParentID      *uuid.UUID      `json:"parent_id"`
	CreatedAt     time.Time       `json:"created_at"`

	// Enriched fields (not in DB table directly, but useful for domain)
	ReplyCount int `json:"reply_count,omitempty" gorm:"-"`
	Depth      int `json:"depth,omitempty" gorm:"-"`
}

const (
	// Thread Types
	ThreadTypeSupport = "support"
	ThreadTypeDirect  = "direct"
	ThreadTypeGroup   = "group"

	// Sender Types
	SenderTypeAgent      = "agent" // Deprecated: Use SenderTypeUser
	SenderTypeUser       = "user"  // For internal members
	SenderTypeSubscriber = "contact"
	SenderTypeSystem     = "system"

	// Message Types
	MessageTypeStandard     = "standard"
	MessageTypeInternalNote = "internal_note"
)

func NewSubscriber(envID uuid.UUID, key string) *Subscriber {
	return &Subscriber{
		ID:            uuid.New(),
		EnvironmentID: envID,
		SubscriberKey: key,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func NewThread(envID uuid.UUID, threadType, channel string) *Thread {
	return &Thread{
		ID:            uuid.New(),
		EnvironmentID: envID,
		Type:          threadType,
		Channel:       channel,
		Status:        "unassigned",
		Metadata:      json.RawMessage("{}"),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func NewThreadParticipant(threadID uuid.UUID, entityType string, entityID uuid.UUID) *ThreadParticipant {
	return &ThreadParticipant{
		ID:         uuid.New(),
		ThreadID:   threadID,
		EntityType: entityType,
		EntityID:   entityID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func NewMessage(tenantID, envID, threadID uuid.UUID, senderType string, senderID *uuid.UUID, content []byte) *Message {
	return &Message{
		ID:            uuid.New(),
		TenantID:      tenantID,
		EnvironmentID: envID,
		ThreadID:      threadID,
		SenderType:    senderType,
		SenderID:      senderID,
		Content:       content,
		Type:          MessageTypeStandard,
		CreatedAt:     time.Now(),
	}
}

func NewInternalNote(tenantID, envID, threadID uuid.UUID, authorID uuid.UUID, content []byte) *Message {
	return &Message{
		ID:            uuid.New(),
		TenantID:      tenantID,
		EnvironmentID: envID,
		ThreadID:      threadID,
		SenderType:    "agent", // Internal notes are always by agents/users
		SenderID:      &authorID,
		Content:       content,
		Type:          MessageTypeInternalNote,
		CreatedAt:     time.Now(),
	}
}

type AssignmentLog struct {
	ID                  uuid.UUID  `json:"id"`
	ThreadID            uuid.UUID  `json:"thread_id"`
	AssignedToMemberID  *uuid.UUID `json:"assigned_to_member_id"`
	AssignedAt          time.Time  `json:"assigned_at"`
	ResolvedAt          *time.Time `json:"resolved_at,omitempty"`
	ResponseTimeSeconds *int       `json:"response_time_seconds,omitempty"`
}

func NewAssignmentLog(threadID uuid.UUID, memberID *uuid.UUID) *AssignmentLog {
	return &AssignmentLog{
		ID:                 uuid.New(),
		ThreadID:           threadID,
		AssignedToMemberID: memberID,
		AssignedAt:         time.Now(),
	}
}

type TeamStats struct {
	TotalConversations       int                      `json:"total_conversations"`
	AvgResponseTime          float64                  `json:"avg_response_time"`
	ResolvedCount            int                      `json:"resolved_count"`
	SLAComplianceRate        float64                  `json:"sla_compliance_rate"`
	ResponseTimeDistribution ResponseTimeDistribution `json:"response_time_distribution"`
}

type ResponseTimeDistribution struct {
	Under5m    int `json:"under_5m"`
	From5To15m int `json:"from_5_to_15m"`
	Over15m    int `json:"over_15m"`
}

type AgentStats struct {
	MemberID                 uuid.UUID                `json:"member_id"`
	TotalAssigned            int                      `json:"total_assigned"`
	TotalResolved            int                      `json:"total_resolved"`
	AvgResponseTime          float64                  `json:"avg_response_time"`
	CurrentOpenThreads       int                      `json:"current_open_threads"`
	Workload                 int                      `json:"workload"` // Same as CurrentOpenThreads for now
	ResponseTimeDistribution ResponseTimeDistribution `json:"response_time_distribution"`
}
