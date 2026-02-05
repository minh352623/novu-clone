package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Role represents a role with permissions
type Role struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Slug        string                 `json:"slug"`
	Permissions map[string]interface{} `json:"permissions"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Role errors
var (
	ErrRoleNameRequired = errors.New("role name is required")
	ErrRoleSlugRequired = errors.New("role slug is required")
)

// Predefined role slugs
const (
	RoleTenantAdmin  = "tenant_admin"
	RoleTenantMember = "tenant_member"
)

// NewRole creates a new role
func NewRole(name, slug string) (*Role, error) {
	if name == "" {
		return nil, ErrRoleNameRequired
	}
	if slug == "" {
		return nil, ErrRoleSlugRequired
	}

	now := time.Now()
	return &Role{
		ID:          uuid.New(),
		Name:        name,
		Slug:        slug,
		Permissions: make(map[string]interface{}),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Validate validates the role
func (r *Role) Validate() error {
	if r.Name == "" {
		return ErrRoleNameRequired
	}
	if r.Slug == "" {
		return ErrRoleSlugRequired
	}
	return nil
}

// SetPermission sets a permission key-value pair
func (r *Role) SetPermission(key string, value interface{}) {
	if r.Permissions == nil {
		r.Permissions = make(map[string]interface{})
	}
	r.Permissions[key] = value
	r.UpdatedAt = time.Now()
}

// HasPermission checks if the role has a specific permission
func (r *Role) HasPermission(key string) bool {
	if r.Permissions == nil {
		return false
	}
	val, exists := r.Permissions[key]
	if !exists {
		return false
	}
	// Check if it's a boolean true
	if b, ok := val.(bool); ok {
		return b
	}
	return exists
}

// RemovePermission removes a permission
func (r *Role) RemovePermission(key string) {
	if r.Permissions != nil {
		delete(r.Permissions, key)
		r.UpdatedAt = time.Now()
	}
}

// IsTenantAdmin checks if this is the tenant admin role
func (r *Role) IsTenantAdmin() bool {
	return r.Slug == RoleTenantAdmin
}
