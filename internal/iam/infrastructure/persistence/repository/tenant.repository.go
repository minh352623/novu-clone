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

type tenantRepository struct {
	db     *gorm.DB
	mapper *mapper.TenantMapper
}

// NewTenantRepository creates a new repository
func NewTenantRepository(db *gorm.DB) repository.TenantRepository {
	return &tenantRepository{
		db:     db,
		mapper: mapper.NewTenantMapper(),
	}
}

func (r *tenantRepository) Create(ctx context.Context, tenant *entity.Tenant) (*entity.Tenant, error) {
	m := r.mapper.ToModel(tenant)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}
	return r.mapper.ToDomain(m), nil
}

func (r *tenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	var m model.TenantModel
	if err := r.db.WithContext(ctx).Preload("PricingPlan").Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	return r.mapper.ToDomain(&m), nil
}

func (r *tenantRepository) GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	var m model.TenantModel
	if err := r.db.WithContext(ctx).Preload("PricingPlan").Where("slug = ?", slug).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	return r.mapper.ToDomain(&m), nil
}

func (r *tenantRepository) GetAll(ctx context.Context, filters repository.TenantFilters) ([]*entity.Tenant, error) {
	var models []model.TenantModel

	query := r.db.WithContext(ctx).Preload("PricingPlan")

	if filters.PricingPlanID != nil {
		query = query.Where("pricing_plan_id = ?", *filters.PricingPlanID)
	}
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
		return nil, fmt.Errorf("failed to get tenants: %w", err)
	}

	tenants := make([]*entity.Tenant, 0, len(models))
	for _, m := range models {
		tenants = append(tenants, r.mapper.ToDomain(&m))
	}

	return tenants, nil
}

func (r *tenantRepository) CountAll(ctx context.Context, filters repository.TenantFilters) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&model.TenantModel{})

	if filters.PricingPlanID != nil {
		query = query.Where("pricing_plan_id = ?", *filters.PricingPlanID)
	}
	if filters.Search != nil && *filters.Search != "" {
		searchPattern := "%" + *filters.Search + "%"
		query = query.Where("name ILIKE ? OR slug ILIKE ?", searchPattern, searchPattern)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	return count, nil
}

func (r *tenantRepository) Update(ctx context.Context, tenant *entity.Tenant) error {
	m := r.mapper.ToModel(tenant)
	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}
	return nil
}

func (r *tenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.TenantModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}
	return nil
}

func (r *tenantRepository) AssignPricingPlan(ctx context.Context, tenantID, planID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&model.TenantModel{}).
		Where("id = ?", tenantID).
		Updates(map[string]interface{}{
			"pricing_plan_id": planID,
			"plan_start_date": gorm.Expr("NOW()"),
		}).Error; err != nil {
		return fmt.Errorf("failed to assign pricing plan: %w", err)
	}
	return nil
}

func (r *tenantRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Tenant, error) {
	var models []model.TenantModel

	if err := r.db.WithContext(ctx).
		Preload("PricingPlan").
		Joins("JOIN tenant_members ON tenant_members.tenant_id = tenants.id").
		Where("tenant_members.user_id = ?", userID).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get tenants by user: %w", err)
	}

	tenants := make([]*entity.Tenant, 0, len(models))
	for _, m := range models {
		tenants = append(tenants, r.mapper.ToDomain(&m))
	}

	return tenants, nil
}
