package impl_test

import (
	"context"
	"testing"
	"time"

	"CONVERDA/internal/health/dto"
	"CONVERDA/internal/health/service/impl"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

type MockQueueReader struct{ mock.Mock }

func (m *MockQueueReader) CountByStatus(ctx context.Context, envID uuid.UUID, status string) (int64, error) {
	args := m.Called(ctx, envID, status)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQueueReader) CountOverdue(ctx context.Context, envID uuid.UUID) (int64, error) {
	args := m.Called(ctx, envID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQueueReader) AvgWaitTime(ctx context.Context, envID uuid.UUID) (float64, error) {
	args := m.Called(ctx, envID)
	return args.Get(0).(float64), args.Error(1)
}

type MockSLAReader struct{ mock.Mock }

func (m *MockSLAReader) GetSLACompliance(ctx context.Context, envID uuid.UUID, slaSeconds int, from, to time.Time) (*dto.SLACompliance, error) {
	args := m.Called(ctx, envID, slaSeconds, from, to)
	return args.Get(0).(*dto.SLACompliance), args.Error(1)
}

type MockWebhookReader struct{ mock.Mock }

func (m *MockWebhookReader) GetHealthStats(ctx context.Context, from, to time.Time) (*dto.WebhookHealth, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).(*dto.WebhookHealth), args.Error(1)
}

// --- Tests ---

func TestHealthService_GetSystemHealth(t *testing.T) {
	t.Run("aggregates all health metrics", func(t *testing.T) {
		qr := new(MockQueueReader)
		sr := new(MockSLAReader)
		wr := new(MockWebhookReader)

		envID := uuid.New()
		from := time.Now().Add(-24 * time.Hour)
		to := time.Now()

		// Queue expectations
		qr.On("CountByStatus", mock.Anything, envID, "unassigned").Return(int64(5), nil)
		qr.On("CountByStatus", mock.Anything, envID, "assigned").Return(int64(10), nil)
		qr.On("CountOverdue", mock.Anything, envID).Return(int64(2), nil)
		qr.On("AvgWaitTime", mock.Anything, envID).Return(120.5, nil)

		// SLA expectations
		sr.On("GetSLACompliance", mock.Anything, envID, 900, from, to).Return(&dto.SLACompliance{
			TotalResolved:  100,
			WithinSLA:      85,
			Breached:       15,
			ComplianceRate: 85.0,
		}, nil)

		// Webhook expectations
		wr.On("GetHealthStats", mock.Anything, from, to).Return(&dto.WebhookHealth{
			TotalDispatched: 50,
			SuccessCount:    45,
			FailedCount:     3,
			PendingRetries:  2,
			SuccessRate:     90.0,
		}, nil)

		svc := impl.NewHealthService(qr, sr, wr, 900)
		result, err := svc.GetSystemHealth(context.Background(), envID, from, to)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Queue
		assert.Equal(t, 5, result.QueueHealth.UnassignedCount)
		assert.Equal(t, 10, result.QueueHealth.AssignedCount)
		assert.Equal(t, 2, result.QueueHealth.OverdueCount)
		assert.Equal(t, 120.5, result.QueueHealth.AvgWaitTimeSec)
		assert.Equal(t, 15, result.QueueHealth.TotalActive) // 5 + 10

		// SLA
		assert.Equal(t, 100, result.SLACompliance.TotalResolved)
		assert.Equal(t, 85, result.SLACompliance.WithinSLA)
		assert.Equal(t, 15, result.SLACompliance.Breached)
		assert.Equal(t, 85.0, result.SLACompliance.ComplianceRate)

		// Webhook
		assert.Equal(t, 50, result.WebhookHealth.TotalDispatched)
		assert.Equal(t, 45, result.WebhookHealth.SuccessCount)
		assert.Equal(t, 90.0, result.WebhookHealth.SuccessRate)

		// GeneratedAt should be recent
		assert.WithinDuration(t, time.Now(), result.GeneratedAt, 2*time.Second)

		qr.AssertExpectations(t)
		sr.AssertExpectations(t)
		wr.AssertExpectations(t)
	})
}
