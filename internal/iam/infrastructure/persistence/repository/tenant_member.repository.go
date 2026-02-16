package repository

import (
	"context"
	"fmt"

	"CONVERDA/internal/iam/domain/model/entity"
	"CONVERDA/internal/iam/domain/repository"
	"CONVERDA/internal/iam/infrastructure/persistence/mapper"
	"CONVERDA/internal/iam/infrastructure/persistence/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type tenantMemberRepository struct {
	db     *gorm.DB
	mapper *mapper.TenantMemberMapper
}

// NewTenantMemberRepository creates a new repository
func NewTenantMemberRepository(db *gorm.DB) repository.TenantMemberRepository {
	return &tenantMemberRepository{
		db:     db,
		mapper: mapper.NewTenantMemberMapper(),
	}
}

func (r *tenantMemberRepository) Create(ctx context.Context, member *entity.TenantMember) (*entity.TenantMember, error) {
	m := r.mapper.ToModel(member)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, fmt.Errorf("failed to create tenant member: %w", err)
	}
	return r.mapper.ToDomain(m), nil
}

func (r *tenantMemberRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.TenantMember, error) {
	var m model.TenantMemberModel
	if err := r.db.WithContext(ctx).
		Preload("Tenant").
		Preload("User").
		Preload("Role").
		Where("id = ?", id).
		First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tenant member: %w", err)
	}
	return r.mapper.ToDomain(&m), nil
}

func (r *tenantMemberRepository) GetByTenantAndUser(ctx context.Context, tenantID, userID uuid.UUID) (*entity.TenantMember, error) {
	// Default to tenant-level check (app_id IS NULL) if not specified?
	// Or should this return ANY membership?
	// Based on previous unique constraint, this returned the single membership.
	// Now there might be multiple.
	// Let's assume this method is for tenant-level membership (where AppID is NULL).
	var m model.TenantMemberModel
	if err := r.db.WithContext(ctx).
		Preload("Tenant").
		Preload("User").
		Preload("Role").
		Where("tenant_id = ? AND user_id = ? AND app_id IS NULL", tenantID, userID).
		First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tenant member: %w", err)
	}
	return r.mapper.ToDomain(&m), nil
}

func (r *tenantMemberRepository) GetAll(ctx context.Context, filters repository.TenantMemberFilters) ([]*entity.TenantMember, error) {
	var models []model.TenantMemberModel

	query := r.db.WithContext(ctx).Preload("Tenant").Preload("User").Preload("Role")

	if filters.TenantID != nil {
		query = query.Where("tenant_id = ?", *filters.TenantID)
	}
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}
	if filters.RoleID != nil {
		query = query.Where("role_id = ?", *filters.RoleID)
	}
	// Add filter for AppID if you add it to filters struct later
	// For now, let's just return everything matching the other filters

	query = query.Order("created_at DESC")

	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get tenant members: %w", err)
	}

	members := make([]*entity.TenantMember, 0, len(models))
	for _, m := range models {
		members = append(members, r.mapper.ToDomain(&m))
	}

	return members, nil
}

func (r *tenantMemberRepository) CountAll(ctx context.Context, filters repository.TenantMemberFilters) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&model.TenantMemberModel{})

	if filters.TenantID != nil {
		query = query.Where("tenant_id = ?", *filters.TenantID)
	}
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}
	if filters.RoleID != nil {
		query = query.Where("role_id = ?", *filters.RoleID)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count tenant members: %w", err)
	}

	return count, nil
}

func (r *tenantMemberRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantMember, error) {
	var models []model.TenantMemberModel

	if err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Role").
		Where("tenant_id = ?", tenantID).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get tenant members: %w", err)
	}

	members := make([]*entity.TenantMember, 0, len(models))
	for _, m := range models {
		members = append(members, r.mapper.ToDomain(&m))
	}

	return members, nil
}

func (r *tenantMemberRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.TenantMember, error) {
	var models []model.TenantMemberModel

	if err := r.db.WithContext(ctx).
		Preload("Tenant").
		Preload("Role").
		Where("user_id = ?", userID).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get memberships: %w", err)
	}

	members := make([]*entity.TenantMember, 0, len(models))
	for _, m := range models {
		members = append(members, r.mapper.ToDomain(&m))
	}

	return members, nil
}

func (r *tenantMemberRepository) Update(ctx context.Context, member *entity.TenantMember) error {
	m := r.mapper.ToModel(member)
	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		return fmt.Errorf("failed to update tenant member: %w", err)
	}
	return nil
}

func (r *tenantMemberRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.TenantMemberModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete tenant member: %w", err)
	}
	return nil
}

func (r *tenantMemberRepository) AssignRole(ctx context.Context, memberID, roleID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&model.TenantMemberModel{}).
		Where("id = ?", memberID).
		Update("role_id", roleID).Error; err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}
	return nil
}

func (r *tenantMemberRepository) RemoveRole(ctx context.Context, memberID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&model.TenantMemberModel{}).
		Where("id = ?", memberID).
		Update("role_id", nil).Error; err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}
	return nil
}

func (r *tenantMemberRepository) ExistsByTenantAndUser(ctx context.Context, tenantID, userID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.TenantMemberModel{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check membership existence: %w", err)
	}
	return count > 0, nil
}

func (r *tenantMemberRepository) GetByTenantUserAndApp(ctx context.Context, tenantID, userID uuid.UUID, appID *uuid.UUID) (*entity.TenantMember, error) {
	var m model.TenantMemberModel
	query := r.db.WithContext(ctx).
		Preload("Tenant").
		Preload("User").
		Preload("Role").
		Where("tenant_id = ? AND user_id = ?", tenantID, userID)

	if appID != nil {
		query = query.Where("app_id = ?", *appID)
	} else {
		query = query.Where("app_id IS NULL")
	}

	if err := query.First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tenant member: %w", err)
	}
	return r.mapper.ToDomain(&m), nil
}
