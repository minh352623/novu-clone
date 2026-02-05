package repository

import (
	"context"

	"CONVERDA/internal/iam/domain/model/entity"

	"github.com/google/uuid"
)

// UserFilters defines filter options for querying users
type UserFilters struct {
	IsRootAdmin *bool
	Search      *string // Search in email and full_name
	Limit       int
	Offset      int
}

// UserRepository defines the interface for user persistence
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *entity.User) (*entity.User, error)

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*entity.User, error)

	// GetAll retrieves all users with filters
	GetAll(ctx context.Context, filters UserFilters) ([]*entity.User, error)

	// CountAll counts all users with filters
	CountAll(ctx context.Context, filters UserFilters) (int64, error)

	// Update updates a user
	Update(ctx context.Context, user *entity.User) error

	// Delete deletes a user
	Delete(ctx context.Context, id uuid.UUID) error

	// UpdateLastLogin updates the last login timestamp
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error

	// UpdatePassword updates the user's password hash
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error

	// ExistsByEmail checks if a user with the given email exists
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
