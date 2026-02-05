package repository

import (
	"context"
	"errors"

	"CONVERDA/internal/iam/domain/model/entity"
	"CONVERDA/internal/iam/domain/repository"
	"CONVERDA/internal/iam/infrastructure/persistence/mapper"
	"CONVERDA/internal/iam/infrastructure/persistence/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type invitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) repository.InvitationRepository {
	return &invitationRepository{db: db}
}

func (r *invitationRepository) Create(ctx context.Context, invitation *entity.TenantInvitation) (*entity.TenantInvitation, error) {
	m := mapper.ToInvitationModel(invitation)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return mapper.ToInvitationEntity(m), nil
}

func (r *invitationRepository) GetByToken(ctx context.Context, token string) (*entity.TenantInvitation, error) {
	var m model.TenantInvitationModel
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil if not found
		}
		return nil, err
	}
	return mapper.ToInvitationEntity(&m), nil
}

func (r *invitationRepository) GetByEmailAndTenant(ctx context.Context, email string, tenantID uuid.UUID) (*entity.TenantInvitation, error) {
	var m model.TenantInvitationModel
	if err := r.db.WithContext(ctx).Where("email = ? AND tenant_id = ? AND status = ?", email, tenantID, entity.InvitationStatusPending).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToInvitationEntity(&m), nil
}

func (r *invitationRepository) Update(ctx context.Context, invitation *entity.TenantInvitation) error {
	m := mapper.ToInvitationModel(invitation)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *invitationRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantInvitation, error) {
	var models []*model.TenantInvitationModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&models).Error; err != nil {
		return nil, err
	}
	return mapper.ToInvitationEntities(models), nil
}
