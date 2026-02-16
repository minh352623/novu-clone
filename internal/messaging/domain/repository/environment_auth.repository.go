package repository

import (
	"context"

	"github.com/google/uuid"
)

type EnvironmentAuthRepository interface {
	ValidateKey(ctx context.Context, apiKey string) (uuid.UUID, uuid.UUID, error) // envID, tenantID
}
