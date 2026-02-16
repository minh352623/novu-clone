package impl

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"CONVERDA/internal/apps/application/service"
	"CONVERDA/internal/apps/domain/model/entity"
	"CONVERDA/internal/apps/domain/repository"

	"github.com/google/uuid"
)

type apiKeyServiceImpl struct {
	apiKeyRepo repository.APIKeyRepository
	envRepo    repository.EnvironmentRepository
}

func NewAPIKeyService(apiKeyRepo repository.APIKeyRepository, envRepo repository.EnvironmentRepository) service.APIKeyService {
	return &apiKeyServiceImpl{
		apiKeyRepo: apiKeyRepo,
		envRepo:    envRepo,
	}
}

func (s *apiKeyServiceImpl) GenerateKey(ctx context.Context, envID uuid.UUID, name string) (string, *entity.APIKey, error) {
	// 1. Get Environment to determine prefix
	env, err := s.envRepo.GetByID(ctx, envID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get environment: %w", err)
	}
	if env == nil {
		return "", nil, fmt.Errorf("environment not found")
	}

	prefix := "sk_test_"
	if env.EnvironmentCode == "prod" || env.EnvironmentCode == "production" {
		prefix = "sk_live_"
	}

	// 2. Generate Random Key
	randomBytes := make([]byte, 24)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	secret := hex.EncodeToString(randomBytes)
	plainKey := prefix + secret

	// 3. Hash the key
	hash := sha256.Sum256([]byte(plainKey))
	keyHash := hex.EncodeToString(hash[:])

	// 4. Create Entity
	suffix := ""
	if len(plainKey) > 4 {
		suffix = plainKey[len(plainKey)-4:]
	}

	apiKey := entity.NewAPIKey(env.AppID, envID, name, prefix, suffix, keyHash, nil)

	// 5. Persist
	createdKey, err := s.apiKeyRepo.Create(ctx, apiKey)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create api key: %w", err)
	}

	return plainKey, createdKey, nil
}

func (s *apiKeyServiceImpl) RotateKey(ctx context.Context, envID uuid.UUID, name string) (string, *entity.APIKey, error) {
	// 1. Revoke existing active keys for this environment
	if err := s.apiKeyRepo.RevokeAllByEnvironment(ctx, envID); err != nil {
		return "", nil, fmt.Errorf("failed to revoke old keys: %w", err)
	}

	// 2. Generate new key
	return s.GenerateKey(ctx, envID, name)
}

func (s *apiKeyServiceImpl) ValidateKey(ctx context.Context, plainKey string) (*entity.APIKey, error) {
	// 1. Hash the incoming key
	hash := sha256.Sum256([]byte(plainKey))
	keyHash := hex.EncodeToString(hash[:])

	// 2. Lookup by hash
	apiKey, err := s.apiKeyRepo.GetByHash(ctx, keyHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get api key by hash: %w", err)
	}
	if apiKey == nil {
		return nil, fmt.Errorf("invalid api key")
	}

	// 3. Check if active
	if !apiKey.IsActive() {
		return nil, fmt.Errorf("api key is expired or revoked")
	}

	return apiKey, nil
}

func (s *apiKeyServiceImpl) RevokeKey(ctx context.Context, keyID uuid.UUID) error {
	apiKey, err := s.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		return fmt.Errorf("failed to get api key by id: %w", err)
	}
	if apiKey == nil {
		return fmt.Errorf("api key not found")
	}

	apiKey.Revoke()
	return s.apiKeyRepo.Update(ctx, apiKey)
}

func (s *apiKeyServiceImpl) ListKeys(ctx context.Context, envID uuid.UUID) ([]*entity.APIKey, error) {
	keys, err := s.apiKeyRepo.GetByEnvironment(ctx, envID)
	if err != nil {
		return nil, fmt.Errorf("failed to list api keys: %w", err)
	}
	return keys, nil
}
