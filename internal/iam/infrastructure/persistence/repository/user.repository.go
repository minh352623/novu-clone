package repository

import (
	"context"
	"fmt"
	"time"

	"CONVERDA/internal/iam/domain/model/entity"
	"CONVERDA/internal/iam/domain/repository"
	"CONVERDA/internal/iam/infrastructure/persistence/mapper"
	"CONVERDA/internal/iam/infrastructure/persistence/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepository struct {
	db     *gorm.DB
	mapper *mapper.UserMapper
}

// NewUserRepository creates a new repository
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{
		db:     db,
		mapper: mapper.NewUserMapper(),
	}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	m := r.mapper.ToModel(user)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return r.mapper.ToDomain(m), nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var m model.UserModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return r.mapper.ToDomain(&m), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var m model.UserModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return r.mapper.ToDomain(&m), nil
}

func (r *userRepository) GetAll(ctx context.Context, filters repository.UserFilters) ([]*entity.User, error) {
	var models []model.UserModel

	query := r.db.WithContext(ctx)

	if filters.IsRootAdmin != nil {
		query = query.Where("is_root_admin = ?", *filters.IsRootAdmin)
	}
	if filters.Search != nil && *filters.Search != "" {
		searchPattern := "%" + *filters.Search + "%"
		query = query.Where("email ILIKE ? OR full_name ILIKE ?", searchPattern, searchPattern)
	}

	query = query.Order("created_at DESC")

	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	users := make([]*entity.User, 0, len(models))
	for _, m := range models {
		users = append(users, r.mapper.ToDomain(&m))
	}

	return users, nil
}

func (r *userRepository) CountAll(ctx context.Context, filters repository.UserFilters) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&model.UserModel{})

	if filters.IsRootAdmin != nil {
		query = query.Where("is_root_admin = ?", *filters.IsRootAdmin)
	}
	if filters.Search != nil && *filters.Search != "" {
		searchPattern := "%" + *filters.Search + "%"
		query = query.Where("email ILIKE ? OR full_name ILIKE ?", searchPattern, searchPattern)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	m := r.mapper.ToModel(user)
	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.UserModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&model.UserModel{}).
		Where("id = ?", id).
		Update("last_login_at", now).Error; err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	if err := r.db.WithContext(ctx).Model(&model.UserModel{}).
		Where("id = ?", id).
		Update("password_hash", passwordHash).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.UserModel{}).
		Where("email = ?", email).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return count > 0, nil
}
