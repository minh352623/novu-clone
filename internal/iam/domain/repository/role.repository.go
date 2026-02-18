package repository

import (
	"context"

	"CONVERDA/internal/iam/domain/model/entity"

	"github.com/google/uuid"
)

// RoleFilters defines filter options for querying roles
type RoleFilters struct {
	Search *string
	Limit  int
	Offset int
}

// RoleReader defines read operations for roles
type RoleReader interface {
	// GetByID retrieves a role by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error)

	// GetBySlug retrieves a role by slug
	GetBySlug(ctx context.Context, slug string) (*entity.Role, error)

	// GetAll retrieves all roles with filters
	GetAll(ctx context.Context, filters RoleFilters) ([]*entity.Role, error)

	// CountAll counts all roles with filters
	CountAll(ctx context.Context, filters RoleFilters) (int64, error)
}

// RoleWriter defines write operations for roles
type RoleWriter interface {
	// Create creates a new role
	Create(ctx context.Context, role *entity.Role) (*entity.Role, error)

	// Update updates a role
	Update(ctx context.Context, role *entity.Role) error

	// Delete deletes a role
	Delete(ctx context.Context, id uuid.UUID) error

	// UpdatePermissions updates the permissions JSONB for a role
	UpdatePermissions(ctx context.Context, id uuid.UUID, permissions map[string]interface{}) error
}

// RoleRepository defines the interface for role persistence by combining reader and writer
type RoleRepository interface {
	RoleReader
	RoleWriter
}
