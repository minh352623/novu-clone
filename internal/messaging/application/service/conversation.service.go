package service

import (
	"context"
	"time"

	"CONVERDA/internal/messaging/controller/dto"
	"CONVERDA/internal/messaging/domain/model/entity"

	"github.com/google/uuid"
)

type ConversationService interface {
	// Inbound (from Webhook/SDK)
	// Inbound (from Webhook/SDK)
	ReceiveMessage(ctx context.Context, tenantID, envID uuid.UUID, subKey, channel string, content []byte, senderID *uuid.UUID) (*entity.Message, error)

	// Outbound (Agent Reply & Notes)
	ReplyMessage(ctx context.Context, tenantID, envID uuid.UUID, threadID uuid.UUID, senderID uuid.UUID, content []byte) (*entity.Message, error)
	AddInternalNote(ctx context.Context, tenantID, envID uuid.UUID, threadID uuid.UUID, authorID uuid.UUID, content []byte) (*entity.Message, error)

	// Thread operations
	GetThread(ctx context.Context, envID, threadID uuid.UUID) (*entity.Thread, error)
	GetMessagesByThread(ctx context.Context, envID, threadID uuid.UUID, limit, offset int) ([]*entity.Message, int64, error)
	GetMessagesByCursor(ctx context.Context, envID, threadID uuid.UUID, cursor string, direction string, limit int) ([]*entity.Message, string, string, error)
	GetOrCreateDirectThread(ctx context.Context, envID, memberID uuid.UUID, targetID uuid.UUID, targetType string) (*entity.Thread, error)
	CreateGroupThread(ctx context.Context, envID uuid.UUID, name string, participants []*entity.ThreadParticipant) (*entity.Thread, error)
	UpdateGroupThread(ctx context.Context, envID, threadID uuid.UUID, name string) error
	AddGroupParticipants(ctx context.Context, envID, threadID uuid.UUID, participants []*entity.ThreadParticipant) error
	RemoveGroupParticipant(ctx context.Context, envID, threadID uuid.UUID, entityType string, entityID uuid.UUID) error
	MarkThreadRead(ctx context.Context, envID, threadID, memberID uuid.UUID) error

	// Assignment
	AssignThread(ctx context.Context, tenantID, envID uuid.UUID, threadID uuid.UUID, memberID uuid.UUID) error
	UnassignThread(ctx context.Context, tenantID, envID uuid.UUID, threadID uuid.UUID) error
	BulkAssignThreads(ctx context.Context, tenantID, envID uuid.UUID, threadIDs []uuid.UUID, memberID uuid.UUID) error
	ResolveThread(ctx context.Context, tenantID, envID uuid.UUID, threadID uuid.UUID) error

	// Analytics & Dashboards
	GetTeamDashboard(ctx context.Context, req dto.DashboardStatsRequest) (*dto.TeamDashboardResponse, error)
	GetPartnerDashboard(ctx context.Context, req dto.DashboardStatsRequest) (*dto.PartnerDashboardResponse, error)
	GetAgentDashboard(ctx context.Context, memberID uuid.UUID, req dto.DashboardStatsRequest) (*dto.AgentDashboardResponse, error)

	// Deprecated: Use Dashboard methods
	GetTeamStats(ctx context.Context, tenantID, envID uuid.UUID, from, to time.Time) (*entity.TeamStats, error)
	GetAgentStats(ctx context.Context, tenantID, envID uuid.UUID, memberID uuid.UUID, from, to time.Time) (*entity.AgentStats, error)

	// Listing
	ListThreads(ctx context.Context, tenantID, envID uuid.UUID, status string, assignedToMe bool, memberID uuid.UUID, limit, offset int) ([]*entity.Thread, int64, error)
}
