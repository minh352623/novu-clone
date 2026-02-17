package impl

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"CONVERDA/global"
	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/pkg/logger"
	"CONVERDA/pkg/setting"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks
type MockWebhookRepository struct {
	mock.Mock
}

func (m *MockWebhookRepository) GetByEvent(ctx context.Context, envID uuid.UUID, eventType string) ([]*entity.Webhook, error) {
	args := m.Called(ctx, envID, eventType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Webhook), args.Error(1)
}

type MockWebhookLogRepository struct {
	mock.Mock
}

func (m *MockWebhookLogRepository) Create(ctx context.Context, log *entity.WebhookLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockWebhookLogRepository) Update(ctx context.Context, log *entity.WebhookLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockWebhookLogRepository) GetPendingRetries(ctx context.Context, limit int) ([]*entity.WebhookLog, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WebhookLog), args.Error(1)
}

func setupTestLogger() {
	global.Logger = logger.NewLogger(setting.LoggerSetting{
		LogLevel: "debug",
	})
}

func TestWebhookDispatcher_Dispatch(t *testing.T) {
	setupTestLogger()

	// 1. Setup Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Headers
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test.event", r.Header.Get("X-Converda-Event"))
		assert.NotEmpty(t, r.Header.Get("X-Converda-Delivery"))

		// Verify Body
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "bar", body["foo"])

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	// 2. Setup Mocks
	mockRepo := new(MockWebhookRepository)
	mockLogRepo := new(MockWebhookLogRepository)

	envID := uuid.New()
	webhook := &entity.Webhook{
		ID:            uuid.New(),
		TenantID:      uuid.New(),
		EnvironmentID: envID,
		URL:           server.URL,
		Events:        []string{"test.event"},
		IsActive:      true,
	}

	mockRepo.On("GetByEvent", mock.Anything, envID, "test.event").Return([]*entity.Webhook{webhook}, nil)
	mockLogRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.WebhookLog")).Return(nil)
	mockLogRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.WebhookLog")).Return(nil)

	// 3. Execute
	dispatcher := NewWebhookDispatcher(mockRepo, mockLogRepo)
	dispatcher.Dispatch(context.Background(), envID, "test.event", map[string]string{"foo": "bar"})

	// Allow goroutine to finish
	time.Sleep(200 * time.Millisecond)

	// 4. Verify
	mockRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}

func TestWebhookDispatcher_FailureSchedulesRetry(t *testing.T) {
	setupTestLogger()

	// Server that always returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	mockRepo := new(MockWebhookRepository)
	mockLogRepo := new(MockWebhookLogRepository)

	envID := uuid.New()
	webhook := &entity.Webhook{
		ID:            uuid.New(),
		TenantID:      uuid.New(),
		EnvironmentID: envID,
		URL:           server.URL,
		Events:        []string{"test.event"},
		IsActive:      true,
	}

	mockRepo.On("GetByEvent", mock.Anything, envID, "test.event").Return([]*entity.Webhook{webhook}, nil)
	mockLogRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.WebhookLog")).Return(nil)
	mockLogRepo.On("Update", mock.Anything, mock.MatchedBy(func(log *entity.WebhookLog) bool {
		// Verify retry is scheduled
		return log.Status == entity.WebhookLogStatusFailed &&
			log.RetryCount == 1 &&
			log.NextRetryAt != nil
	})).Return(nil)

	dispatcher := NewWebhookDispatcher(mockRepo, mockLogRepo)
	dispatcher.Dispatch(context.Background(), envID, "test.event", map[string]string{"foo": "bar"})

	time.Sleep(200 * time.Millisecond)

	mockRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}
