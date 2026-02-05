package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"CONVERDA/internal/iam/domain/model/entity"
	"CONVERDA/internal/iam/domain/repository"
	"CONVERDA/internal/iam/infrastructure/persistence/mapper"
	"CONVERDA/internal/iam/infrastructure/persistence/model"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type roleRepository struct {
	db     *gorm.DB
	mapper *mapper.RoleMapper
}

// NewRoleRepository creates a new repository
func NewRoleRepository(db *gorm.DB) repository.RoleRepository {
	return &roleRepository{
		db:     db,
		mapper: mapper.NewRoleMapper(),
	}
}

func (r *roleRepository) Create(ctx context.Context, role *entity.Role) (*entity.Role, error) {
	m := r.mapper.ToModel(role)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}
	return r.mapper.ToDomain(m), nil
}

func (r *roleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	var m model.RoleModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return r.mapper.ToDomain(&m), nil
}

func (r *roleRepository) GetBySlug(ctx context.Context, slug string) (*entity.Role, error) {
	var m model.RoleModel
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return r.mapper.ToDomain(&m), nil
}

func (r *roleRepository) GetAll(ctx context.Context, filters repository.RoleFilters) ([]*entity.Role, error) {
	var models []model.RoleModel

	query := r.db.WithContext(ctx)

	if filters.Search != nil && *filters.Search != "" {
		searchPattern := "%" + *filters.Search + "%"
		query = query.Where("name ILIKE ? OR slug ILIKE ?", searchPattern, searchPattern)
	}

	query = query.Order("created_at DESC")

	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	roles := make([]*entity.Role, 0, len(models))
	for _, m := range models {
		roles = append(roles, r.mapper.ToDomain(&m))
	}

	return roles, nil
}

func (r *roleRepository) CountAll(ctx context.Context, filters repository.RoleFilters) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&model.RoleModel{})

	if filters.Search != nil && *filters.Search != "" {
		searchPattern := "%" + *filters.Search + "%"
		query = query.Where("name ILIKE ? OR slug ILIKE ?", searchPattern, searchPattern)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count roles: %w", err)
	}

	return count, nil
}

func (r *roleRepository) Update(ctx context.Context, role *entity.Role) error {
	m := r.mapper.ToModel(role)
	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}
	return nil
}

func (r *roleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.RoleModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}

func (r *roleRepository) UpdatePermissions(ctx context.Context, id uuid.UUID, permissions map[string]interface{}) error {
	permissionsJSON, err := json.Marshal(permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	if err := r.db.WithContext(ctx).Model(&model.RoleModel{}).
		Where("id = ?", id).
		Update("permissions", datatypes.JSON(permissionsJSON)).Error; err != nil {
		return fmt.Errorf("failed to update permissions: %w", err)
	}
	return nil
}
