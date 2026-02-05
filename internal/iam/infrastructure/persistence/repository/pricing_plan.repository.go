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

type pricingPlanRepository struct {
	db     *gorm.DB
	mapper *mapper.PricingPlanMapper
}

// NewPricingPlanRepository creates a new repository
func NewPricingPlanRepository(db *gorm.DB) repository.PricingPlanRepository {
	return &pricingPlanRepository{
		db:     db,
		mapper: mapper.NewPricingPlanMapper(),
	}
}

func (r *pricingPlanRepository) Create(ctx context.Context, plan *entity.PricingPlan) (*entity.PricingPlan, error) {
	m := r.mapper.ToModel(plan)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, fmt.Errorf("failed to create pricing plan: %w", err)
	}
	return r.mapper.ToDomain(m), nil
}

func (r *pricingPlanRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.PricingPlan, error) {
	var m model.PricingPlanModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get pricing plan: %w", err)
	}
	return r.mapper.ToDomain(&m), nil
}

func (r *pricingPlanRepository) GetBySlug(ctx context.Context, slug string) (*entity.PricingPlan, error) {
	var m model.PricingPlanModel
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get pricing plan: %w", err)
	}
	return r.mapper.ToDomain(&m), nil
}

func (r *pricingPlanRepository) GetDefault(ctx context.Context) (*entity.PricingPlan, error) {
	var m model.PricingPlanModel
	if err := r.db.WithContext(ctx).Where("is_default = ? AND is_active = ?", true, true).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get default pricing plan: %w", err)
	}
	return r.mapper.ToDomain(&m), nil
}

func (r *pricingPlanRepository) GetAll(ctx context.Context, filters repository.PricingPlanFilters) ([]*entity.PricingPlan, error) {
	var models []model.PricingPlanModel

	query := r.db.WithContext(ctx)

	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}
	if filters.IsDefault != nil {
		query = query.Where("is_default = ?", *filters.IsDefault)
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
		return nil, fmt.Errorf("failed to get pricing plans: %w", err)
	}

	plans := make([]*entity.PricingPlan, 0, len(models))
	for _, m := range models {
		plans = append(plans, r.mapper.ToDomain(&m))
	}

	return plans, nil
}

func (r *pricingPlanRepository) CountAll(ctx context.Context, filters repository.PricingPlanFilters) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&model.PricingPlanModel{})

	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}
	if filters.IsDefault != nil {
		query = query.Where("is_default = ?", *filters.IsDefault)
	}
	if filters.Search != nil && *filters.Search != "" {
		searchPattern := "%" + *filters.Search + "%"
		query = query.Where("name ILIKE ? OR slug ILIKE ?", searchPattern, searchPattern)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count pricing plans: %w", err)
	}

	return count, nil
}

func (r *pricingPlanRepository) Update(ctx context.Context, plan *entity.PricingPlan) error {
	m := r.mapper.ToModel(plan)
	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		return fmt.Errorf("failed to update pricing plan: %w", err)
	}
	return nil
}

func (r *pricingPlanRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.PricingPlanModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete pricing plan: %w", err)
	}
	return nil
}

func (r *pricingPlanRepository) SetAsDefault(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Unset all defaults
		if err := tx.Model(&model.PricingPlanModel{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return fmt.Errorf("failed to unset defaults: %w", err)
		}
		// Set new default
		if err := tx.Model(&model.PricingPlanModel{}).Where("id = ?", id).Update("is_default", true).Error; err != nil {
			return fmt.Errorf("failed to set default: %w", err)
		}
		return nil
	})
}
