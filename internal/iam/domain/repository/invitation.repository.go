package repository

import (
	"context"

	"CONVERDA/internal/iam/domain/model/entity"

	"github.com/google/uuid"
)

type InvitationRepository interface {
	Create(ctx context.Context, invitation *entity.TenantInvitation) (*entity.TenantInvitation, error)
	GetByToken(ctx context.Context, token string) (*entity.TenantInvitation, error)
	GetByEmailAndTenant(ctx context.Context, email string, tenantID uuid.UUID) (*entity.TenantInvitation, error)
	Update(ctx context.Context, invitation *entity.TenantInvitation) error
	// Maybe list pending invitations for a tenant
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantInvitation, error)
}
