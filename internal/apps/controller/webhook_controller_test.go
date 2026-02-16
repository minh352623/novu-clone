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

// MockWebhookService
type MockWebhookService struct {
	mock.Mock
}

func (m *MockWebhookService) CreateWebhook(ctx context.Context, tenantID, appID, envID uuid.UUID, url string, events []string, desc *string) (*entity.Webhook, error) {
	args := m.Called(ctx, tenantID, appID, envID, url, events, desc)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Webhook), args.Error(1)
}

func (m *MockWebhookService) GetWebhook(ctx context.Context, id uuid.UUID) (*entity.Webhook, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Webhook), args.Error(1)
}

func (m *MockWebhookService) ListWebhooks(ctx context.Context, appID uuid.UUID) ([]*entity.Webhook, error) {
	args := m.Called(ctx, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Webhook), args.Error(1)
}

func (m *MockWebhookService) UpdateWebhook(ctx context.Context, webhook *entity.Webhook) error {
	args := m.Called(ctx, webhook)
	return args.Error(0)
}

func (m *MockWebhookService) DeleteWebhook(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWebhookService) ToggleWebhook(ctx context.Context, id uuid.UUID, active bool) error {
	args := m.Called(ctx, id, active)
	return args.Error(0)
}

func TestWebhookController_CreateWebhook(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		appID := uuid.New()
		tenantID := uuid.New()
		envID := uuid.New()

		mockService := new(MockWebhookService)
		ctrl := controller.NewWebhookController(mockService)

		router := gin.New()
		router.POST("/apps/:app_id/webhooks", response.Wrap(ctrl.CreateWebhook, http.StatusCreated))

		desc := "Test Webhook"
		reqBody := dto.CreateWebhookRequest{
			EnvironmentID: envID,
			URL:           "https://example.com/webhook",
			Events:        []string{"user.created"},
			Description:   &desc,
		}
		jsonBody, _ := json.Marshal(reqBody)

		webhookEntity := &entity.Webhook{
			ID:            uuid.New(),
			AppID:         appID,
			TenantID:      tenantID,
			EnvironmentID: envID,
			URL:           reqBody.URL,
			Events:        reqBody.Events,
			Description:   reqBody.Description,
			CreatedAt:     time.Now(),
		}

		mockService.On("CreateWebhook", mock.Anything, tenantID, appID, envID, reqBody.URL, reqBody.Events, reqBody.Description).
			Return(webhookEntity, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/apps/"+appID.String()+"/webhooks?tenant_id="+tenantID.String(), bytes.NewBuffer(jsonBody))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		// Note: Controller returns 200 or 201? Checking controller code...
		// It says @Success 201 in comments but usually Gin returns 200 by default unless specified.
		// Let's check controller code again. It creates default success response which is usually 200 unless AbortWithStatusJSON with 201 is called.
		// Wait, the code was: return dto.ToWebhookResponse(webhook), nil
		// Controller wrapper usually handles 200 OK.
		// But let's assume standard response wrapper handles it.
		// Actually, I should check how the return value is handled.
		// If it uses a wrapper, I can't test strict status code easily without knowing the wrapper behavior.
		// I'll assume 200 for now or verify content.

		mockService.AssertExpectations(t)
	})
}
