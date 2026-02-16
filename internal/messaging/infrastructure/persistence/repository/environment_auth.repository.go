package repository

import (
	"context"

	domainRepo "CONVERDA/internal/messaging/domain/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type environmentAuthRepository struct {
	db *gorm.DB
}

func NewEnvironmentAuthRepository(db *gorm.DB) domainRepo.EnvironmentAuthRepository {
	return &environmentAuthRepository{db: db}
}

func (r *environmentAuthRepository) ValidateKey(ctx context.Context, apiKey string) (uuid.UUID, uuid.UUID, error) {
	var result struct {
		EnvironmentID uuid.UUID `gorm:"column:environment_id"`
		TenantID      uuid.UUID `gorm:"column:tenant_id"`
	}

	// Join environments -> apps -> tenants to get both IDs
	err := r.db.WithContext(ctx).Raw(`
		SELECT e.id as environment_id, a.tenant_id
		FROM environments e
		JOIN apps a ON e.app_id = a.id
		WHERE e.api_key = ?
		LIMIT 1
	`, apiKey).Scan(&result).Error

	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	if result.EnvironmentID == uuid.Nil {
		return uuid.Nil, uuid.Nil, gorm.ErrRecordNotFound
	}

	return result.EnvironmentID, result.TenantID, nil
}
