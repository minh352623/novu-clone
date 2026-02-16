package impl_test

import (
	"context"
	"testing"
	"time"

	"CONVERDA/internal/apps/application/service/impl"
	"CONVERDA/internal/apps/domain/model/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAPIKeyRepository is a mock implementation
type MockAPIKeyRepository struct {
	mock.Mock
}

func (m *MockAPIKeyRepository) Create(ctx context.Context, apiKey *entity.APIKey) (*entity.APIKey, error) {
	args := m.Called(ctx, apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.APIKey), args.Error(1)
}

func (m *MockAPIKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.APIKey, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.APIKey), args.Error(1)
}

func (m *MockAPIKeyRepository) GetByHash(ctx context.Context, hash string) (*entity.APIKey, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.APIKey), args.Error(1)
}

func (m *MockAPIKeyRepository) GetByEnvironment(ctx context.Context, envID uuid.UUID) ([]*entity.APIKey, error) {
	args := m.Called(ctx, envID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.APIKey), args.Error(1)
}

func (m *MockAPIKeyRepository) Update(ctx context.Context, apiKey *entity.APIKey) error {
	args := m.Called(ctx, apiKey)
	return args.Error(0)
}

func (m *MockAPIKeyRepository) RevokeAllByEnvironment(ctx context.Context, envID uuid.UUID) error {
	args := m.Called(ctx, envID)
	return args.Error(0)
}

// MockEnvironmentRepository is a mock implementation
type MockEnvironmentRepository struct {
	mock.Mock
}

func (m *MockEnvironmentRepository) Create(ctx context.Context, env *entity.Environment) (*entity.Environment, error) {
	args := m.Called(ctx, env)
	return args.Get(0).(*entity.Environment), args.Error(1)
}

func (m *MockEnvironmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Environment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Environment), args.Error(1)
}

func (m *MockEnvironmentRepository) GetByAppAndCode(ctx context.Context, appID uuid.UUID, code string) (*entity.Environment, error) {
	args := m.Called(ctx, appID, code)
	return args.Get(0).(*entity.Environment), args.Error(1)
}

func (m *MockEnvironmentRepository) ListByApp(ctx context.Context, appID uuid.UUID) ([]*entity.Environment, error) {
	args := m.Called(ctx, appID)
	return args.Get(0).([]*entity.Environment), args.Error(1)
}

func (m *MockEnvironmentRepository) GetByAPIKey(ctx context.Context, apiKey string) (*entity.Environment, error) {
	args := m.Called(ctx, apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Environment), args.Error(1)
}

func (m *MockEnvironmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestAPIKeyService_GenerateKey(t *testing.T) {
	ctx := context.Background()
	envID := uuid.New()
	appID := uuid.New()

	mockAPIKeyRepo := new(MockAPIKeyRepository)
	mockEnvRepo := new(MockEnvironmentRepository)

	service := impl.NewAPIKeyService(mockAPIKeyRepo, mockEnvRepo)

	t.Run("successful generation for prod environment", func(t *testing.T) {
		env := &entity.Environment{
			ID:              envID,
			AppID:           appID,
			EnvironmentCode: "prod",
		}

		mockEnvRepo.On("GetByID", ctx, envID).Return(env, nil).Once()
		mockAPIKeyRepo.On("Create", ctx, mock.AnythingOfType("*entity.APIKey")).Return(&entity.APIKey{
			ID:            uuid.New(),
			KeyPrefix:     "sk_live_",
			EnvironmentID: envID,
		}, nil).Once()

		plainKey, createdKey, err := service.GenerateKey(ctx, envID, "My Key")

		assert.NoError(t, err)
		assert.NotNil(t, createdKey)
		assert.Contains(t, plainKey, "sk_live_")
		assert.Equal(t, "sk_live_", createdKey.KeyPrefix)
		mockEnvRepo.AssertExpectations(t)
		mockAPIKeyRepo.AssertExpectations(t)
	})

	t.Run("environment not found", func(t *testing.T) {
		mockEnvRepo.On("GetByID", ctx, envID).Return(nil, nil).Once()

		plainKey, createdKey, err := service.GenerateKey(ctx, envID, "My Key")

		assert.Error(t, err)
		assert.Empty(t, plainKey)
		assert.Nil(t, createdKey)
		assert.Contains(t, err.Error(), "environment not found")
	})
}

func TestAPIKeyService_ValidateKey(t *testing.T) {
	ctx := context.Background()
	plainKey := "sk_test_1234567890abcdef"

	mockAPIKeyRepo := new(MockAPIKeyRepository)
	mockEnvRepo := new(MockEnvironmentRepository)
	service := impl.NewAPIKeyService(mockAPIKeyRepo, mockEnvRepo)

	t.Run("valid key", func(t *testing.T) {
		apiKey := &entity.APIKey{
			ID:        uuid.New(),
			KeyPrefix: "sk_test_",
			RevokedAt: nil,
		}

		mockAPIKeyRepo.On("GetByHash", ctx, mock.Anything).Return(apiKey, nil).Once()

		validKey, err := service.ValidateKey(ctx, plainKey)

		assert.NoError(t, err)
		assert.NotNil(t, validKey)
		assert.Equal(t, apiKey.ID, validKey.ID)
	})

	t.Run("revoked key", func(t *testing.T) {
		revokedAt := time.Now()
		apiKey := &entity.APIKey{
			ID:        uuid.New(),
			KeyPrefix: "sk_test_",
			RevokedAt: &revokedAt,
		}

		mockAPIKeyRepo.On("GetByHash", ctx, mock.Anything).Return(apiKey, nil).Once()

		validKey, err := service.ValidateKey(ctx, plainKey)

		assert.Error(t, err)
		assert.Nil(t, validKey)
		assert.Contains(t, err.Error(), "expired or revoked")
	})
}
