package worker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"CONVERDA/global"
	appsEntity "CONVERDA/internal/apps/domain/model/entity"
	"CONVERDA/internal/messaging/application/worker"
	"CONVERDA/internal/messaging/domain/model/entity"
	domainRepo "CONVERDA/internal/messaging/domain/repository"
	"CONVERDA/pkg/logger"
	"CONVERDA/pkg/setting"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

type MockThreadRepository struct {
	mock.Mock
}

func (m *MockThreadRepository) Create(ctx context.Context, thread *entity.Thread) (*entity.Thread, error) {
	args := m.Called(ctx, thread)
	return args.Get(0).(*entity.Thread), args.Error(1)
}

func (m *MockThreadRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Thread, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.Thread), args.Error(1)
}

func (m *MockThreadRepository) GetByIDAndEnv(ctx context.Context, id, envID uuid.UUID) (*entity.Thread, error) {
	args := m.Called(ctx, id, envID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Thread), args.Error(1)
}

func (m *MockThreadRepository) Update(ctx context.Context, thread *entity.Thread) error {
	args := m.Called(ctx, thread)
	return args.Error(0)
}

func (m *MockThreadRepository) List(ctx context.Context, filter domainRepo.ThreadFilter) ([]*entity.Thread, int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*entity.Thread), args.Get(1).(int64), args.Error(2)
}

func (m *MockThreadRepository) AddParticipant(ctx context.Context, p *entity.ThreadParticipant) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockThreadRepository) UpdateParticipant(ctx context.Context, p *entity.ThreadParticipant) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockThreadRepository) RemoveParticipant(ctx context.Context, threadID uuid.UUID, entityType string, entityID uuid.UUID) error {
	args := m.Called(ctx, threadID, entityType, entityID)
	return args.Error(0)
}

func (m *MockThreadRepository) GetParticipants(ctx context.Context, threadID uuid.UUID) ([]*entity.ThreadParticipant, error) {
	args := m.Called(ctx, threadID)
	return args.Get(0).([]*entity.ThreadParticipant), args.Error(1)
}

func (m *MockThreadRepository) GetDirectThreadBetweenEntities(ctx context.Context, typeA string, idA uuid.UUID, typeB string, idB uuid.UUID) (*entity.Thread, error) {
	args := m.Called(ctx, typeA, idA, typeB, idB)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Thread), args.Error(1)
}

func (m *MockThreadRepository) MarkAsOverdue(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockAssignmentLogRepository struct {
	mock.Mock
}

func (m *MockAssignmentLogRepository) Create(ctx context.Context, log *entity.AssignmentLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAssignmentLogRepository) GetLastByThread(ctx context.Context, threadID uuid.UUID) (*entity.AssignmentLog, error) {
	args := m.Called(ctx, threadID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AssignmentLog), args.Error(1)
}

func (m *MockAssignmentLogRepository) Update(ctx context.Context, log *entity.AssignmentLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAssignmentLogRepository) GetTeamStats(ctx context.Context, envID uuid.UUID, from, to time.Time, slaSeconds int) (*entity.TeamStats, error) {
	args := m.Called(ctx, envID, from, to, slaSeconds)
	return args.Get(0).(*entity.TeamStats), args.Error(1)
}

func (m *MockAssignmentLogRepository) GetAgentStats(ctx context.Context, envID uuid.UUID, memberID uuid.UUID, from, to time.Time, slaSeconds int) (*entity.AgentStats, error) {
	args := m.Called(ctx, envID, memberID, from, to, slaSeconds)
	return args.Get(0).(*entity.AgentStats), args.Error(1)
}

func (m *MockAssignmentLogRepository) GetActivityTimeline(ctx context.Context, envID, memberID uuid.UUID, from, to time.Time, interval string) ([]*entity.ActivityPoint, error) {
	args := m.Called(ctx, envID, memberID, from, to, interval)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.ActivityPoint), args.Error(1)
}

func (m *MockAssignmentLogRepository) GetByThread(ctx context.Context, threadID uuid.UUID) ([]*entity.AssignmentLog, error) {
	args := m.Called(ctx, threadID)
	return args.Get(0).([]*entity.AssignmentLog), args.Error(1)
}

type MockAppRepository struct {
	mock.Mock
}

func (m *MockAppRepository) Create(ctx context.Context, app *appsEntity.App) (*appsEntity.App, error) {
	args := m.Called(ctx, app)
	return args.Get(0).(*appsEntity.App), args.Error(1)
}

func (m *MockAppRepository) GetByID(ctx context.Context, id uuid.UUID) (*appsEntity.App, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*appsEntity.App), args.Error(1)
}

func (m *MockAppRepository) Update(ctx context.Context, app *appsEntity.App) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockAppRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAppRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*appsEntity.App, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]*appsEntity.App), args.Error(1)
}

type MockEnvironmentRepository struct {
	mock.Mock
}

func (m *MockEnvironmentRepository) Create(ctx context.Context, env *appsEntity.Environment) (*appsEntity.Environment, error) {
	args := m.Called(ctx, env)
	return args.Get(0).(*appsEntity.Environment), args.Error(1)
}

func (m *MockEnvironmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*appsEntity.Environment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*appsEntity.Environment), args.Error(1)
}

func (m *MockEnvironmentRepository) GetByAppAndCode(ctx context.Context, appID uuid.UUID, code string) (*appsEntity.Environment, error) {
	args := m.Called(ctx, appID, code)
	return args.Get(0).(*appsEntity.Environment), args.Error(1)
}

func (m *MockEnvironmentRepository) ListByApp(ctx context.Context, appID uuid.UUID) ([]*appsEntity.Environment, error) {
	args := m.Called(ctx, appID)
	return args.Get(0).([]*appsEntity.Environment), args.Error(1)
}

func (m *MockEnvironmentRepository) GetByAPIKey(ctx context.Context, apiKey string) (*appsEntity.Environment, error) {
	args := m.Called(ctx, apiKey)
	return args.Get(0).(*appsEntity.Environment), args.Error(1)
}

func (m *MockEnvironmentRepository) Update(ctx context.Context, env *appsEntity.Environment) error {
	args := m.Called(ctx, env)
	return args.Error(0)
}

func (m *MockEnvironmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// --- Helper ---

func setupLogger() {
	global.Logger = logger.NewLogger(setting.LoggerSetting{
		LogLevel: "debug",
	})
}

func newSLAWorker(threadRepo *MockThreadRepository, logRepo *MockAssignmentLogRepository, appRepo *MockAppRepository, envRepo *MockEnvironmentRepository) *worker.SLAWorker {
	return worker.NewSLAWorker(threadRepo, logRepo, appRepo, envRepo)
}

// Matcher for ThreadFilter by status
func statusMatcher(status string) interface{} {
	return mock.MatchedBy(func(f domainRepo.ThreadFilter) bool {
		return f.Status != nil && *f.Status == status
	})
}

// --- Tests ---

func TestSLAWorker_CheckSLA(t *testing.T) {
	setupLogger()

	t.Run("should mark unassigned thread as overdue", func(t *testing.T) {
		ctx := context.Background()
		mockThreadRepo := new(MockThreadRepository)
		mockLogRepo := new(MockAssignmentLogRepository)
		mockAppRepo := new(MockAppRepository)
		mockEnvRepo := new(MockEnvironmentRepository)
		w := newSLAWorker(mockThreadRepo, mockLogRepo, mockAppRepo, mockEnvRepo)

		envID := uuid.New()
		threadID := uuid.New()

		mockEnvRepo.On("GetByID", ctx, envID).Return(&appsEntity.Environment{
			ID:                  envID,
			SLAThresholdSeconds: 10,
		}, nil).Once()

		thread := &entity.Thread{
			ID:            threadID,
			EnvironmentID: envID,
			Status:        "unassigned",
			CreatedAt:     time.Now().Add(-15 * time.Second),
			IsOverdue:     false,
		}

		mockThreadRepo.On("List", ctx, statusMatcher("unassigned")).Return([]*entity.Thread{thread}, int64(1), nil).Once()
		mockThreadRepo.On("List", ctx, statusMatcher("assigned")).Return([]*entity.Thread{}, int64(0), nil).Once()
		mockThreadRepo.On("MarkAsOverdue", ctx, threadID).Return(nil).Once()

		w.Exported_checkSLA(ctx)

		mockThreadRepo.AssertExpectations(t)
		mockEnvRepo.AssertExpectations(t)
	})

	t.Run("should mark assigned thread as overdue based on assignment time", func(t *testing.T) {
		ctx := context.Background()
		mockThreadRepo := new(MockThreadRepository)
		mockLogRepo := new(MockAssignmentLogRepository)
		mockAppRepo := new(MockAppRepository)
		mockEnvRepo := new(MockEnvironmentRepository)
		w := newSLAWorker(mockThreadRepo, mockLogRepo, mockAppRepo, mockEnvRepo)

		envID := uuid.New()
		threadID := uuid.New()
		memberID := uuid.New()

		mockEnvRepo.On("GetByID", ctx, envID).Return(&appsEntity.Environment{
			ID:                  envID,
			SLAThresholdSeconds: 10,
		}, nil).Once()

		thread := &entity.Thread{
			ID:            threadID,
			EnvironmentID: envID,
			Status:        "assigned",
			CreatedAt:     time.Now().Add(-1 * time.Hour),
			IsOverdue:     false,
		}

		mockThreadRepo.On("List", ctx, statusMatcher("unassigned")).Return([]*entity.Thread{}, int64(0), nil).Once()
		mockThreadRepo.On("List", ctx, statusMatcher("assigned")).Return([]*entity.Thread{thread}, int64(1), nil).Once()
		mockLogRepo.On("GetLastByThread", ctx, threadID).Return(&entity.AssignmentLog{
			ThreadID:           threadID,
			AssignedToMemberID: &memberID,
			AssignedAt:         time.Now().Add(-15 * time.Second),
		}, nil).Once()
		mockThreadRepo.On("MarkAsOverdue", ctx, threadID).Return(nil).Once()

		w.Exported_checkSLA(ctx)

		mockThreadRepo.AssertExpectations(t)
		mockEnvRepo.AssertExpectations(t)
		mockLogRepo.AssertExpectations(t)
	})

	t.Run("should NOT mark thread within SLA threshold", func(t *testing.T) {
		ctx := context.Background()
		mockThreadRepo := new(MockThreadRepository)
		mockLogRepo := new(MockAssignmentLogRepository)
		mockAppRepo := new(MockAppRepository)
		mockEnvRepo := new(MockEnvironmentRepository)
		w := newSLAWorker(mockThreadRepo, mockLogRepo, mockAppRepo, mockEnvRepo)

		envID := uuid.New()

		// SLA = 60s, thread created 5s ago — NOT overdue
		mockEnvRepo.On("GetByID", ctx, envID).Return(&appsEntity.Environment{
			ID:                  envID,
			SLAThresholdSeconds: 60,
		}, nil).Once()

		thread := &entity.Thread{
			ID:            uuid.New(),
			EnvironmentID: envID,
			Status:        "unassigned",
			CreatedAt:     time.Now().Add(-5 * time.Second),
			IsOverdue:     false,
		}

		mockThreadRepo.On("List", ctx, statusMatcher("unassigned")).Return([]*entity.Thread{thread}, int64(1), nil).Once()
		mockThreadRepo.On("List", ctx, statusMatcher("assigned")).Return([]*entity.Thread{}, int64(0), nil).Once()
		// MarkAsOverdue should NOT be called

		w.Exported_checkSLA(ctx)

		mockThreadRepo.AssertExpectations(t)
		mockThreadRepo.AssertNotCalled(t, "MarkAsOverdue", mock.Anything, mock.Anything)
	})

	t.Run("should use default 900s when env not found", func(t *testing.T) {
		ctx := context.Background()
		mockThreadRepo := new(MockThreadRepository)
		mockLogRepo := new(MockAssignmentLogRepository)
		mockAppRepo := new(MockAppRepository)
		mockEnvRepo := new(MockEnvironmentRepository)
		w := newSLAWorker(mockThreadRepo, mockLogRepo, mockAppRepo, mockEnvRepo)

		envID := uuid.New()

		// Env lookup fails → default 900s
		mockEnvRepo.On("GetByID", ctx, envID).Return(nil, errors.New("not found")).Once()

		// Thread created 5 seconds ago — well under 900s default
		thread := &entity.Thread{
			ID:            uuid.New(),
			EnvironmentID: envID,
			Status:        "unassigned",
			CreatedAt:     time.Now().Add(-5 * time.Second),
			IsOverdue:     false,
		}

		mockThreadRepo.On("List", ctx, statusMatcher("unassigned")).Return([]*entity.Thread{thread}, int64(1), nil).Once()
		mockThreadRepo.On("List", ctx, statusMatcher("assigned")).Return([]*entity.Thread{}, int64(0), nil).Once()

		w.Exported_checkSLA(ctx)

		mockThreadRepo.AssertNotCalled(t, "MarkAsOverdue", mock.Anything, mock.Anything)
	})

	t.Run("should fallback to App SLA when Env SLA is 0", func(t *testing.T) {
		ctx := context.Background()
		mockThreadRepo := new(MockThreadRepository)
		mockLogRepo := new(MockAssignmentLogRepository)
		mockAppRepo := new(MockAppRepository)
		mockEnvRepo := new(MockEnvironmentRepository)
		w := newSLAWorker(mockThreadRepo, mockLogRepo, mockAppRepo, mockEnvRepo)

		envID := uuid.New()
		appID := uuid.New()
		threadID := uuid.New()

		// Env SLA = 0, App SLA = 10
		mockEnvRepo.On("GetByID", ctx, envID).Return(&appsEntity.Environment{
			ID:                  envID,
			AppID:               appID,
			SLAThresholdSeconds: 0, // No env-level SLA
		}, nil).Once()
		mockAppRepo.On("GetByID", ctx, appID).Return(&appsEntity.App{
			ID:                  appID,
			SLAThresholdSeconds: 10, // App-level SLA = 10s
		}, nil).Once()

		// Thread created 15s ago → should be overdue (10s threshold from App)
		thread := &entity.Thread{
			ID:            threadID,
			EnvironmentID: envID,
			Status:        "unassigned",
			CreatedAt:     time.Now().Add(-15 * time.Second),
			IsOverdue:     false,
		}

		mockThreadRepo.On("List", ctx, statusMatcher("unassigned")).Return([]*entity.Thread{thread}, int64(1), nil).Once()
		mockThreadRepo.On("List", ctx, statusMatcher("assigned")).Return([]*entity.Thread{}, int64(0), nil).Once()
		mockThreadRepo.On("MarkAsOverdue", ctx, threadID).Return(nil).Once()

		w.Exported_checkSLA(ctx)

		mockThreadRepo.AssertExpectations(t)
		mockAppRepo.AssertExpectations(t)
	})

	t.Run("should handle empty thread list without panic", func(t *testing.T) {
		ctx := context.Background()
		mockThreadRepo := new(MockThreadRepository)
		mockLogRepo := new(MockAssignmentLogRepository)
		mockAppRepo := new(MockAppRepository)
		mockEnvRepo := new(MockEnvironmentRepository)
		w := newSLAWorker(mockThreadRepo, mockLogRepo, mockAppRepo, mockEnvRepo)

		mockThreadRepo.On("List", ctx, statusMatcher("unassigned")).Return([]*entity.Thread{}, int64(0), nil).Once()
		mockThreadRepo.On("List", ctx, statusMatcher("assigned")).Return([]*entity.Thread{}, int64(0), nil).Once()

		// Should not panic — just complete cleanly
		w.Exported_checkSLA(ctx)

		mockThreadRepo.AssertExpectations(t)
		mockThreadRepo.AssertNotCalled(t, "MarkAsOverdue", mock.Anything, mock.Anything)
	})

	t.Run("should handle list error gracefully", func(t *testing.T) {
		ctx := context.Background()
		mockThreadRepo := new(MockThreadRepository)
		mockLogRepo := new(MockAssignmentLogRepository)
		mockAppRepo := new(MockAppRepository)
		mockEnvRepo := new(MockEnvironmentRepository)
		w := newSLAWorker(mockThreadRepo, mockLogRepo, mockAppRepo, mockEnvRepo)

		mockThreadRepo.On("List", ctx, statusMatcher("unassigned")).Return([]*entity.Thread{}, int64(0), errors.New("db error")).Once()
		mockThreadRepo.On("List", ctx, statusMatcher("assigned")).Return([]*entity.Thread{}, int64(0), nil).Once()

		// Should not panic — logs error and moves to next status
		w.Exported_checkSLA(ctx)

		mockThreadRepo.AssertExpectations(t)
		mockThreadRepo.AssertNotCalled(t, "MarkAsOverdue", mock.Anything, mock.Anything)
	})

	t.Run("should use default threshold for nil envID", func(t *testing.T) {
		ctx := context.Background()
		mockThreadRepo := new(MockThreadRepository)
		mockLogRepo := new(MockAssignmentLogRepository)
		mockAppRepo := new(MockAppRepository)
		mockEnvRepo := new(MockEnvironmentRepository)
		w := newSLAWorker(mockThreadRepo, mockLogRepo, mockAppRepo, mockEnvRepo)

		// Thread with nil envID, created 5s ago — under 900s default
		thread := &entity.Thread{
			ID:            uuid.New(),
			EnvironmentID: uuid.Nil,
			Status:        "unassigned",
			CreatedAt:     time.Now().Add(-5 * time.Second),
			IsOverdue:     false,
		}

		mockThreadRepo.On("List", ctx, statusMatcher("unassigned")).Return([]*entity.Thread{thread}, int64(1), nil).Once()
		mockThreadRepo.On("List", ctx, statusMatcher("assigned")).Return([]*entity.Thread{}, int64(0), nil).Once()

		w.Exported_checkSLA(ctx)

		// Default is 900s, thread is 5s old → NOT overdue
		mockThreadRepo.AssertNotCalled(t, "MarkAsOverdue", mock.Anything, mock.Anything)
		// envRepo should NOT be called for uuid.Nil
		mockEnvRepo.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
	})
}
