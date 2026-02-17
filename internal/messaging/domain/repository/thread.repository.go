package repository

import (
	"context"

	"CONVERDA/internal/messaging/domain/model/entity"

	"github.com/google/uuid"
)

type ThreadFilter struct {
	Type          *string
	Status        *string
	AssignedToMe  bool
	MemberID      *uuid.UUID
	EnvironmentID *uuid.UUID
	IsOverdue     *bool
	Limit         int
	Offset        int
}

type ThreadRepository interface {
	Create(ctx context.Context, thread *entity.Thread) (*entity.Thread, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Thread, error)
	Update(ctx context.Context, thread *entity.Thread) error
	List(ctx context.Context, filter ThreadFilter) ([]*entity.Thread, int64, error)

	// Participant management
	AddParticipant(ctx context.Context, p *entity.ThreadParticipant) error
	UpdateParticipant(ctx context.Context, p *entity.ThreadParticipant) error
	RemoveParticipant(ctx context.Context, threadID uuid.UUID, entityType string, entityID uuid.UUID) error
	GetParticipants(ctx context.Context, threadID uuid.UUID) ([]*entity.ThreadParticipant, error)

	// Secure management
	GetByIDAndEnv(ctx context.Context, id, envID uuid.UUID) (*entity.Thread, error)

	// Direct Chat Specific
	GetDirectThreadBetweenEntities(ctx context.Context, typeA string, idA uuid.UUID, typeB string, idB uuid.UUID) (*entity.Thread, error)

	MarkAsOverdue(ctx context.Context, id uuid.UUID) error
}
