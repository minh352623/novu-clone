package service

import (
	"context"

	"CONVERDA/internal/apps/domain/model/entity"

	"github.com/google/uuid"
)

type APIKeyService interface {
	// GenerateKey creates a new API key for the environment
	GenerateKey(ctx context.Context, envID uuid.UUID, name string) (string, *entity.APIKey, error)

	// RotateKey invalidates old keys and creates a new one
	RotateKey(ctx context.Context, envID uuid.UUID, name string) (string, *entity.APIKey, error)

	// ValidateKey checks if a plain api key is valid
	ValidateKey(ctx context.Context, plainKey string) (*entity.APIKey, error)

	// RevokeKey revokes a specific key
	RevokeKey(ctx context.Context, keyID uuid.UUID) error

	// ListKeys returns all keys for an environment
	ListKeys(ctx context.Context, envID uuid.UUID) ([]*entity.APIKey, error)
}
