package mapper

import (
	"CONVERDA/internal/iam/domain/model/entity"
	"CONVERDA/internal/iam/infrastructure/persistence/model"
)

// ToInvitationEntity converts model to entity
func ToInvitationEntity(m *model.TenantInvitationModel) *entity.TenantInvitation {
	if m == nil {
		return nil
	}
	return &entity.TenantInvitation{
		ID:        m.ID,
		TenantID:  m.TenantID,
		Email:     m.Email,
		RoleID:    m.RoleID,
		Token:     m.Token,
		Status:    entity.InvitationStatus(m.Status),
		InvitedBy: m.InvitedBy,
		ExpiresAt: m.ExpiresAt,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// ToInvitationModel converts entity to model
func ToInvitationModel(e *entity.TenantInvitation) *model.TenantInvitationModel {
	if e == nil {
		return nil
	}
	return &model.TenantInvitationModel{
		ID:        e.ID,
		TenantID:  e.TenantID,
		Email:     e.Email,
		RoleID:    e.RoleID,
		Token:     e.Token,
		Status:    string(e.Status),
		InvitedBy: e.InvitedBy,
		ExpiresAt: e.ExpiresAt,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

// ToInvitationEntities converts list of models to entities
func ToInvitationEntities(models []*model.TenantInvitationModel) []*entity.TenantInvitation {
	entities := make([]*entity.TenantInvitation, len(models))
	for i, m := range models {
		entities[i] = ToInvitationEntity(m)
	}
	return entities
}
