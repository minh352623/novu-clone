package dto

import (
	"encoding/json"
	"time"

	"CONVERDA/internal/messaging/domain/model/entity"

	"github.com/google/uuid"
)

type InboundMessageRequest struct {
	EnvironmentID uuid.UUID       `json:"environment_id" binding:"required"`
	SubscriberKey string          `json:"subscriber_key" binding:"required"`
	Channel       string          `json:"channel"` // Optional, default "api" or "web"
	Content       json.RawMessage `json:"content" binding:"required" swaggertype:"object"`
}

type ReplyMessageRequest struct {
	Content json.RawMessage `json:"content" binding:"required" swaggertype:"object"`
}

type AssignConversationRequest struct {
	MemberID *uuid.UUID `json:"member_id"` // Optional, if empty assigns to caller
}

type InternalNoteRequest struct {
	Content json.RawMessage `json:"content" binding:"required" swaggertype:"object"`
}

type BulkAssignRequest struct {
	ThreadIDs []uuid.UUID `json:"thread_ids" binding:"required"`
	MemberID  *uuid.UUID  `json:"member_id"` // Optional
}

type WsEvent struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

type ParticipantDTO struct {
	ID   uuid.UUID `json:"id" binding:"required"`
	Type string    `json:"type" binding:"required,oneof=user subscriber"`
}

type CreateDirectChatRequest struct {
	Target ParticipantDTO `json:"target" binding:"required"`
}

type CreateGroupChatRequest struct {
	Name         string           `json:"name" binding:"required"`
	Participants []ParticipantDTO `json:"participants" binding:"required,min=1"`
}

type UpdateGroupChatRequest struct {
	Name string `json:"name" binding:"required"`
}

type AddParticipantsRequest struct {
	Participants []ParticipantDTO `json:"participants" binding:"required,min=1"`
}

type MessageResponse struct {
	ID         uuid.UUID       `json:"id"`
	ThreadID   uuid.UUID       `json:"thread_id"`
	SenderType string          `json:"sender_type"`
	Content    json.RawMessage `json:"content" swaggertype:"object"`
	CreatedAt  time.Time       `json:"created_at"`
	ParentID   *uuid.UUID      `json:"parent_id,omitempty"`
}

type ThreadResponse struct {
	ID            uuid.UUID              `json:"id"`
	EnvironmentID uuid.UUID              `json:"environment_id"`
	Type          string                 `json:"type"`
	Channel       string                 `json:"channel"`
	Status        string                 `json:"status"`
	Metadata      json.RawMessage        `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Participants  []*ParticipantResponse `json:"participants,omitempty"`
}

type ParticipantResponse struct {
	EntityID    uuid.UUID `json:"entity_id"`
	EntityType  string    `json:"entity_type"`
	DisplayName string    `json:"display_name,omitempty"`
}

type TeamStatsResponse struct {
	TotalConversations int     `json:"total_conversations"`
	AvgResponseTime    float64 `json:"avg_response_time"`
	ResolvedCount      int     `json:"resolved_count"`
	SLAComplianceRate  float64 `json:"sla_compliance_rate"`
}

type AgentStatsResponse struct {
	MemberID           uuid.UUID `json:"member_id"`
	TotalAssigned      int       `json:"total_assigned"`
	TotalResolved      int       `json:"total_resolved"`
	AvgResponseTime    float64   `json:"avg_response_time"`
	CurrentOpenThreads int       `json:"current_open_threads"`
}

type AssignmentLogResponse struct {
	ID                    uuid.UUID  `json:"id"`
	ThreadID              uuid.UUID  `json:"thread_id"`
	AssignedToMemberID    *uuid.UUID `json:"assigned_to_member_id"`
	AssignedToDisplayName string     `json:"assigned_to_display_name,omitempty"`
	AssignedAt            time.Time  `json:"assigned_at"`
	ResolvedAt            *time.Time `json:"resolved_at,omitempty"`
	ResponseTimeSeconds   *int       `json:"response_time_seconds,omitempty"`
}

type AuditTrailResponse struct {
	ThreadID uuid.UUID                `json:"thread_id"`
	Logs     []*AssignmentLogResponse `json:"logs"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
}

func ToMessageResponse(m *entity.Message) *MessageResponse {
	return &MessageResponse{
		ID:         m.ID,
		ThreadID:   m.ThreadID,
		SenderType: m.SenderType,
		Content:    m.Content,
		CreatedAt:  m.CreatedAt,
		ParentID:   m.ParentID,
	}
}

func ToMessageResponseList(msgs []*entity.Message) []*MessageResponse {
	res := make([]*MessageResponse, len(msgs))
	for i, m := range msgs {
		res[i] = ToMessageResponse(m)
	}
	return res
}

func ToThreadResponse(p *entity.Thread) *ThreadResponse {
	res := &ThreadResponse{
		ID:            p.ID,
		EnvironmentID: p.EnvironmentID,
		Type:          p.Type,
		Channel:       p.Channel,
		Status:        string(p.Status),
		Metadata:      p.Metadata,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
	for _, part := range p.Participants {
		res.Participants = append(res.Participants, &ParticipantResponse{
			EntityID:    part.EntityID,
			EntityType:  part.EntityType,
			DisplayName: part.DisplayName,
		})
	}
	return res
}

func ToThreadResponseList(threads []*entity.Thread) []*ThreadResponse {
	res := make([]*ThreadResponse, len(threads))
	for i, p := range threads {
		res[i] = ToThreadResponse(p)
	}
	return res
}
