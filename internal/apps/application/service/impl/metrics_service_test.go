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

func (m *MockUsageMetricRepository) GetDetailedSummary(ctx context.Context, appID uuid.UUID, from, to time.Time) (*entity.DetailedMetricsSummary, error) {
	args := m.Called(ctx, appID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.DetailedMetricsSummary), args.Error(1)
}

func (m *MockUsageMetricRepository) GetDailyTimeSeries(ctx context.Context, appID uuid.UUID, envID *uuid.UUID, from, to time.Time) ([]entity.DailyMetric, error) {
	args := m.Called(ctx, appID, envID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.DailyMetric), args.Error(1)
}

func (m *MockUsageMetricRepository) GetEnvironmentBreakdown(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]entity.EnvironmentMetric, error) {
	args := m.Called(ctx, appID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.EnvironmentMetric), args.Error(1)
}

func TestMetricsService_RecordUsage(t *testing.T) {
	ctx := context.Background()
	appID := uuid.New()
	envID := uuid.New()

	mockRepo := new(MockUsageMetricRepository)
	service := impl.NewMetricsService(mockRepo)

	t.Run("successful recording with direction", func(t *testing.T) {
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.UsageMetric")).Return(nil).Once()

		err := service.RecordUsage(ctx, appID, envID, "sms", "outbound", 200)

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

func TestMetricsService_GetDetailedMetrics(t *testing.T) {
	ctx := context.Background()
	appID := uuid.New()

	mockRepo := new(MockUsageMetricRepository)
	service := impl.NewMetricsService(mockRepo)

	t.Run("returns detailed breakdown", func(t *testing.T) {
		summary := &entity.DetailedMetricsSummary{
			TotalMessages:    100,
			InboundMessages:  30,
			OutboundMessages: 70,
			SuccessCount:     85,
			FailureCount:     15,
			ByProvider: map[string]entity.ProviderMetric{
				"email": {Total: 60, Success: 55, Failed: 5},
				"sms":   {Total: 40, Success: 30, Failed: 10},
			},
		}
		mockRepo.On("GetDetailedSummary", ctx, appID, mock.Anything, mock.Anything).Return(summary, nil).Once()

		result, err := service.GetDetailedMetrics(ctx, appID, 30)

		assert.NoError(t, err)
		assert.Equal(t, int64(100), result.TotalMessages)
		assert.Equal(t, int64(30), result.InboundMessages)
		assert.Equal(t, int64(70), result.OutboundMessages)
		assert.InDelta(t, 85.0, result.SuccessRate, 0.01)
		assert.Len(t, result.ByProvider, 2)
		mockRepo.AssertExpectations(t)
	})
}

func TestMetricsService_GetDailyTimeSeries(t *testing.T) {
	ctx := context.Background()
	appID := uuid.New()

	mockRepo := new(MockUsageMetricRepository)
	service := impl.NewMetricsService(mockRepo)

	t.Run("returns daily data", func(t *testing.T) {
		dailyMetrics := []entity.DailyMetric{
			{Date: "2026-02-15", Inbound: 5, Outbound: 10, Success: 13, Failed: 2},
			{Date: "2026-02-16", Inbound: 8, Outbound: 12, Success: 18, Failed: 2},
		}
		mockRepo.On("GetDailyTimeSeries", ctx, appID, (*uuid.UUID)(nil), mock.Anything, mock.Anything).Return(dailyMetrics, nil).Once()

		result, err := service.GetDailyTimeSeries(ctx, appID, nil, 7)

		assert.NoError(t, err)
		assert.Len(t, result.Days, 2)
		assert.Equal(t, "2026-02-15", result.Days[0].Date)
		mockRepo.AssertExpectations(t)
	})
}

func TestMetricsService_GetEnvironmentBreakdown(t *testing.T) {
	ctx := context.Background()
	appID := uuid.New()
	envID := uuid.New()

	mockRepo := new(MockUsageMetricRepository)
	service := impl.NewMetricsService(mockRepo)

	t.Run("returns per-environment data", func(t *testing.T) {
		envMetrics := []entity.EnvironmentMetric{
			{EnvironmentID: envID, Total: 50, Inbound: 20, Outbound: 30, Success: 45, Failed: 5},
		}
		mockRepo.On("GetEnvironmentBreakdown", ctx, appID, mock.Anything, mock.Anything).Return(envMetrics, nil).Once()

		result, err := service.GetEnvironmentBreakdown(ctx, appID, 30)

		assert.NoError(t, err)
		assert.Len(t, result.Environments, 1)
		assert.Equal(t, envID, result.Environments[0].EnvironmentID)
		assert.Equal(t, int64(50), result.Environments[0].Total)
		mockRepo.AssertExpectations(t)
	})
}
