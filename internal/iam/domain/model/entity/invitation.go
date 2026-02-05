package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvitationExpired  = errors.New("invitation has expired")
	ErrInvitationAccepted = errors.New("invitation has already been accepted")
	ErrInvalidToken       = errors.New("invalid invitation token")
)

type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusExpired  InvitationStatus = "expired"
	InvitationStatusDeclined InvitationStatus = "declined"
)

type TenantInvitation struct {
	ID        uuid.UUID        `json:"id"`
	TenantID  uuid.UUID        `json:"tenant_id"`
	Email     string           `json:"email"`
	RoleID    *uuid.UUID       `json:"role_id,omitempty"`
	Token     string           `json:"-"`
	Status    InvitationStatus `json:"status"`
	InvitedBy *uuid.UUID       `json:"invited_by,omitempty"`
	ExpiresAt time.Time        `json:"expires_at"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

func NewTenantInvitation(tenantID uuid.UUID, email string, roleID *uuid.UUID, invitedBy *uuid.UUID, duration time.Duration) (*TenantInvitation, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}

	// Generate secure token (UUID is simple enough for now, or crypto random)
	token := uuid.New().String()

	return &TenantInvitation{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Email:     email,
		RoleID:    roleID,
		Token:     token,
		Status:    InvitationStatusPending,
		InvitedBy: invitedBy,
		ExpiresAt: time.Now().Add(duration),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (i *TenantInvitation) Accept() error {
	if i.Status != InvitationStatusPending {
		return ErrInvitationAccepted
	}
	if time.Now().After(i.ExpiresAt) {
		i.Status = InvitationStatusExpired
		return ErrInvitationExpired
	}
	i.Status = InvitationStatusAccepted
	return nil
}

func (i *TenantInvitation) Validate() error {
	if i.Email == "" {
		return errors.New("email is required")
	}
	if i.Token == "" {
		return errors.New("token is required")
	}
	return nil
}
