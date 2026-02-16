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

// MockUsageMetricRepository is a mock implementation
type MockUsageMetricRepository struct {
	mock.Mock
}

func (m *MockUsageMetricRepository) Create(ctx context.Context, metric *entity.UsageMetric) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockUsageMetricRepository) GetSummaryByApp(ctx context.Context, appID uuid.UUID, from, to time.Time) (map[string]int64, error) {
	args := m.Called(ctx, appID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockUsageMetricRepository) GetEnvironmentUsage(ctx context.Context, envID uuid.UUID, from, to time.Time) (int64, error) {
	args := m.Called(ctx, envID, from, to)
	return args.Get(0).(int64), args.Error(1)
}

func TestMetricsService_RecordUsage(t *testing.T) {
	ctx := context.Background()
	appID := uuid.New()
	envID := uuid.New()

	mockRepo := new(MockUsageMetricRepository)
	service := impl.NewMetricsService(mockRepo)

	t.Run("successful recording", func(t *testing.T) {
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.UsageMetric")).Return(nil).Once()

		err := service.RecordUsage(ctx, appID, envID, "sms", 200)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestMetricsService_GetAppMetrics(t *testing.T) {
	ctx := context.Background()
	appID := uuid.New()

	mockRepo := new(MockUsageMetricRepository)
	service := impl.NewMetricsService(mockRepo)

	t.Run("successful summary retrieval", func(t *testing.T) {
		expectedMetrics := map[string]int64{"success": 10, "error": 2}
		mockRepo.On("GetSummaryByApp", ctx, appID, mock.Anything, mock.Anything).Return(expectedMetrics, nil).Once()

		metrics, err := service.GetAppMetrics(ctx, appID, 7)

		assert.NoError(t, err)
		assert.Equal(t, expectedMetrics, metrics)
		mockRepo.AssertExpectations(t)
	})
}
