package impl

import (
	"context"
	"fmt"

	"CONVERDA/internal/iam/application/service"
	"CONVERDA/internal/iam/domain/model/entity"
	"CONVERDA/internal/iam/domain/repository"

	"github.com/google/uuid"
)

// roleServiceImpl implements RoleService
type roleServiceImpl struct {
	roleRepo repository.RoleRepository
}

// NewRoleService creates a new RoleService
func NewRoleService(roleRepo repository.RoleRepository) service.RoleService {
	return &roleServiceImpl{
		roleRepo: roleRepo,
	}
}

func (s *roleServiceImpl) CreateRole(ctx context.Context, name, slug string, permissions map[string]interface{}) (*entity.Role, error) {
	// Check if slug exists
	existing, err := s.roleRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing role slug: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("role with slug '%s' already exists", slug)
	}

	// Create role
	role, err := entity.NewRole(name, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// Set permissions
	if permissions != nil {
		role.Permissions = permissions
	}

	// Persist
	createdRole, err := s.roleRepo.Create(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("failed to save role: %w", err)
	}

	return createdRole, nil
}

func (s *roleServiceImpl) GetRole(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return nil, service.ErrRoleNotFound
	}
	return role, nil
}

func (s *roleServiceImpl) ListRoles(ctx context.Context, filters repository.RoleFilters) ([]*entity.Role, int64, error) {
	roles, err := s.roleRepo.GetAll(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list roles: %w", err)
	}

	count, err := s.roleRepo.CountAll(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count roles: %w", err)
	}

	return roles, count, nil
}

func (s *roleServiceImpl) UpdateRole(ctx context.Context, role *entity.Role) error {
	if err := role.Validate(); err != nil {
		return err
	}
	if err := s.roleRepo.Update(ctx, role); err != nil {
		return fmt.Errorf("failed to update role %s: %w", role.ID, err)
	}
	return nil
}

func (s *roleServiceImpl) DeleteRole(ctx context.Context, id uuid.UUID) error {
	if err := s.roleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete role %s: %w", id, err)
	}
	return nil
}

func (s *roleServiceImpl) UpdatePermissions(ctx context.Context, roleID uuid.UUID, permissions map[string]interface{}) error {
	if err := s.roleRepo.UpdatePermissions(ctx, roleID, permissions); err != nil {
		return fmt.Errorf("failed to update permissions for role %s: %w", roleID, err)
	}
	return nil
}
