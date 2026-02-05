package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// TenantMember represents a user's membership in a tenant
type TenantMember struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	UserID    uuid.UUID  `json:"user_id"`
	RoleID    *uuid.UUID `json:"role_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	// Relationships (loaded separately)
	Tenant *Tenant `json:"tenant,omitempty"`
	User   *User   `json:"user,omitempty"`
	Role   *Role   `json:"role,omitempty"`
}

// TenantMember errors
var (
	ErrMemberTenantRequired = errors.New("tenant ID is required")
	ErrMemberUserRequired   = errors.New("user ID is required")
	ErrMemberAlreadyExists  = errors.New("user is already a member of this tenant")
)

// NewTenantMember creates a new tenant member
func NewTenantMember(tenantID, userID uuid.UUID, roleID *uuid.UUID) (*TenantMember, error) {
	if tenantID == uuid.Nil {
		return nil, ErrMemberTenantRequired
	}
	if userID == uuid.Nil {
		return nil, ErrMemberUserRequired
	}

	now := time.Now()
	return &TenantMember{
		ID:        uuid.New(),
		TenantID:  tenantID,
		UserID:    userID,
		RoleID:    roleID,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Validate validates the tenant member
func (m *TenantMember) Validate() error {
	if m.TenantID == uuid.Nil {
		return ErrMemberTenantRequired
	}
	if m.UserID == uuid.Nil {
		return ErrMemberUserRequired
	}
	return nil
}

// AssignRole assigns a role to the member
func (m *TenantMember) AssignRole(roleID uuid.UUID) {
	m.RoleID = &roleID
	m.UpdatedAt = time.Now()
}

// RemoveRole removes the role from the member
func (m *TenantMember) RemoveRole() {
	m.RoleID = nil
	m.UpdatedAt = time.Now()
}

// HasRole checks if the member has a role assigned
func (m *TenantMember) HasRole() bool {
	return m.RoleID != nil
}
