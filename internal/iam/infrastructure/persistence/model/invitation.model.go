package model

import (
	"time"

	"github.com/google/uuid"
)

// TenantInvitationModel is the GORM model for tenant_invitations table
type TenantInvitationModel struct {
	ID        uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	TenantID  uuid.UUID  `gorm:"column:tenant_id;type:uuid;not null"`
	Email     string     `gorm:"column:email;type:text;not null;index"`
	RoleID    *uuid.UUID `gorm:"column:role_id;type:uuid"`
	Token     string     `gorm:"column:token;type:text;uniqueIndex;not null"`
	Status    string     `gorm:"column:status;type:text;not null;default:'pending'"`
	InvitedBy *uuid.UUID `gorm:"column:invited_by;type:uuid"`
	ExpiresAt time.Time  `gorm:"column:expires_at;not null"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime"`

	// Relationships
	Tenant  *TenantModel `gorm:"foreignKey:TenantID"`
	Role    *RoleModel   `gorm:"foreignKey:RoleID"`
	Inviter *UserModel   `gorm:"foreignKey:InvitedBy"`
}

func (TenantInvitationModel) TableName() string {
	return "tenant_invitations"
}
