package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"CONVERDA/internal/apps/controller"
	"CONVERDA/internal/apps/controller/dto"
	"CONVERDA/internal/apps/domain/model/entity"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProviderService
type MockProviderService struct {
	mock.Mock
}

func (m *MockProviderService) CreateProvider(ctx context.Context, tenantID, appID, envID uuid.UUID, providerType, providerName string, config map[string]interface{}) (*entity.Provider, error) {
	args := m.Called(ctx, tenantID, appID, envID, providerType, providerName, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Provider), args.Error(1)
}

func (m *MockProviderService) GetProvider(ctx context.Context, id uuid.UUID) (*entity.Provider, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Provider), args.Error(1)
}

func (m *MockProviderService) ListProviders(ctx context.Context, appID uuid.UUID) ([]*entity.Provider, error) {
	args := m.Called(ctx, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Provider), args.Error(1)
}

func (m *MockProviderService) GetActiveProvider(ctx context.Context, envID uuid.UUID, providerType string) (*entity.Provider, error) {
	args := m.Called(ctx, envID, providerType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Provider), args.Error(1)
}

func (m *MockProviderService) UpdateProvider(ctx context.Context, provider *entity.Provider) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

func (m *MockProviderService) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProviderService) ToggleProvider(ctx context.Context, id uuid.UUID, active bool) error {
	args := m.Called(ctx, id, active)
	return args.Error(0)
}

func TestProviderController_CreateProvider(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		appID := uuid.New()
		tenantID := uuid.New()
		envID := uuid.New()

		mockService := new(MockProviderService)
		ctrl := controller.NewProviderController(mockService)

		router := gin.New()
		router.POST("/apps/:app_id/providers", response.Wrap(ctrl.CreateProvider, http.StatusCreated))

		config := map[string]interface{}{
			"apiKey": "test_api_key",
		}
		reqBody := dto.CreateProviderRequest{
			EnvironmentID: envID,
			ProviderType:  "email",
			ProviderName:  "Test Provider",
			Configuration: config,
		}
		jsonBody, _ := json.Marshal(reqBody)

		providerEntity := &entity.Provider{
			ID:            uuid.New(),
			AppID:         appID,
			TenantID:      tenantID,
			EnvironmentID: envID,
			ProviderType:  reqBody.ProviderType,
			ProviderName:  reqBody.ProviderName,
			Configuration: reqBody.Configuration,
			CreatedAt:     time.Now(),
		}

		mockService.On("CreateProvider", mock.Anything, tenantID, appID, envID, reqBody.ProviderType, reqBody.ProviderName, reqBody.Configuration).
			Return(providerEntity, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/apps/"+appID.String()+"/providers?tenant_id="+tenantID.String(), bytes.NewBuffer(jsonBody))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})
}
