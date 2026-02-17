package repository

import (
	"context"

	"github.com/google/uuid"
)

// MemberInfo is a DTO for tenant member data
type MemberInfo struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	UserID      uuid.UUID
	RoleID      *uuid.UUID
	DisplayName string
}

// MemberReader defines the interface for reading tenant member data from other modules
type MemberReader interface {
	GetMember(ctx context.Context, id uuid.UUID) (*MemberInfo, error)
}
