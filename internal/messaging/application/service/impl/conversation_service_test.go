package impl_test

import (
	"context"
	"testing"
	"time"

	appsEntity "CONVERDA/internal/apps/domain/model/entity"
	iamEntity "CONVERDA/internal/iam/domain/model/entity"
	iamRepo "CONVERDA/internal/iam/domain/repository"
	"CONVERDA/internal/messaging/application/service/impl"
	"CONVERDA/internal/messaging/domain/model/entity"
	domainRepo "CONVERDA/internal/messaging/domain/repository"
	"CONVERDA/pkg/cursor"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Create(ctx context.Context, msg *entity.Message) (*entity.Message, error) {
	args := m.Called(ctx, msg)
	return args.Get(0).(*entity.Message), args.Error(1)
}

func (m *MockMessageRepository) GetMessagesByThread(ctx context.Context, threadID uuid.UUID) ([]*entity.Message, error) {
	args := m.Called(ctx, threadID)
	return args.Get(0).([]*entity.Message), args.Error(1)
}

func (m *MockMessageRepository) ListByThread(ctx context.Context, threadID uuid.UUID, limit, offset int) ([]*entity.Message, error) {
	args := m.Called(ctx, threadID, limit, offset)
	return args.Get(0).([]*entity.Message), args.Error(1)
}

func (m *MockMessageRepository) ListByCursor(ctx context.Context, threadID uuid.UUID, cursor *cursor.Cursor, direction string, limit int) ([]*entity.Message, error) {
	args := m.Called(ctx, threadID, cursor, direction, limit)
	return args.Get(0).([]*entity.Message), args.Error(1)
}

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

type MockSubscriberRepository struct {
	mock.Mock
}

func (m *MockSubscriberRepository) Create(ctx context.Context, sub *entity.Subscriber) (*entity.Subscriber, error) {
	args := m.Called(ctx, sub)
	return args.Get(0).(*entity.Subscriber), args.Error(1)
}

func (m *MockSubscriberRepository) GetByKey(ctx context.Context, envID uuid.UUID, key string) (*entity.Subscriber, error) {
	args := m.Called(ctx, envID, key)
	return args.Get(0).(*entity.Subscriber), args.Error(1)
}

func (m *MockSubscriberRepository) Update(ctx context.Context, sub *entity.Subscriber) error {
	args := m.Called(ctx, sub)
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

type MockMemberRepository struct {
	mock.Mock
}

func (m *MockMemberRepository) GetByID(ctx context.Context, id uuid.UUID) (*iamEntity.TenantMember, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*iamEntity.TenantMember), args.Error(1)
}

func (m *MockMemberRepository) Create(ctx context.Context, member *iamEntity.TenantMember) (*iamEntity.TenantMember, error) {
	return nil, nil
}
func (m *MockMemberRepository) GetByTenantAndUser(ctx context.Context, tenantID, userID uuid.UUID) (*iamEntity.TenantMember, error) {
	return nil, nil
}
func (m *MockMemberRepository) GetAll(ctx context.Context, filters iamRepo.TenantMemberFilters) ([]*iamEntity.TenantMember, error) {
	return nil, nil
}
func (m *MockMemberRepository) CountAll(ctx context.Context, filters iamRepo.TenantMemberFilters) (int64, error) {
	return 0, nil
}
func (m *MockMemberRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*iamEntity.TenantMember, error) {
	return nil, nil
}
func (m *MockMemberRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*iamEntity.TenantMember, error) {
	return nil, nil
}
func (m *MockMemberRepository) Update(ctx context.Context, member *iamEntity.TenantMember) error {
	return nil
}
func (m *MockMemberRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *MockMemberRepository) AssignRole(ctx context.Context, memberID, roleID uuid.UUID) error {
	return nil
}
func (m *MockMemberRepository) RemoveRole(ctx context.Context, memberID uuid.UUID) error {
	return nil
}
func (m *MockMemberRepository) ExistsByTenantAndUser(ctx context.Context, tenantID, userID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *MockMemberRepository) GetByTenantUserAndApp(ctx context.Context, tenantID, userID uuid.UUID, appID *uuid.UUID) (*iamEntity.TenantMember, error) {
	return nil, nil
}

type MockAppRepository struct {
	mock.Mock
}

func (m *MockAppRepository) Create(ctx context.Context, app *appsEntity.App) (*appsEntity.App, error) {
	return nil, nil
}
func (m *MockAppRepository) GetByID(ctx context.Context, id uuid.UUID) (*appsEntity.App, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*appsEntity.App), args.Error(1)
}
func (m *MockAppRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*appsEntity.App, error) {
	return nil, nil
}
func (m *MockAppRepository) Update(ctx context.Context, app *appsEntity.App) error {
	return nil
}
func (m *MockAppRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

type MockEnvironmentRepository struct {
	mock.Mock
}

func (m *MockEnvironmentRepository) Create(ctx context.Context, env *appsEntity.Environment) (*appsEntity.Environment, error) {
	return nil, nil
}
func (m *MockEnvironmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*appsEntity.Environment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*appsEntity.Environment), args.Error(1)
}
func (m *MockEnvironmentRepository) GetByAppAndCode(ctx context.Context, appID uuid.UUID, code string) (*appsEntity.Environment, error) {
	return nil, nil
}
func (m *MockEnvironmentRepository) ListByApp(ctx context.Context, appID uuid.UUID) ([]*appsEntity.Environment, error) {
	return nil, nil
}
func (m *MockEnvironmentRepository) GetByAPIKey(ctx context.Context, apiKey string) (*appsEntity.Environment, error) {
	return nil, nil
}
func (m *MockEnvironmentRepository) Update(ctx context.Context, env *appsEntity.Environment) error {
	return nil
}
func (m *MockEnvironmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

// --- Tests ---

func TestConversationService_GetOrCreateDirectThread(t *testing.T) {
	ctx := context.Background()
	envID := uuid.New()
	memberID := uuid.New()
	targetSubID := uuid.New()

	mockMsgRepo := new(MockMessageRepository)
	mockThreadRepo := new(MockThreadRepository)
	mockSubRepo := new(MockSubscriberRepository)
	mockLogRepo := new(MockAssignmentLogRepository)
	mockMemberRepo := new(MockMemberRepository)
	mockAppRepo := new(MockAppRepository)
	mockEnvRepo := new(MockEnvironmentRepository)

	service := impl.NewConversationService(mockMsgRepo, mockThreadRepo, mockSubRepo, mockLogRepo, mockMemberRepo, mockAppRepo, mockEnvRepo, nil)

	t.Run("creates new thread between user and subscriber", func(t *testing.T) {
		// Mock: Create thread succeeds
		mockThreadRepo.On("Create", ctx, mock.AnythingOfType("*entity.Thread")).Return(&entity.Thread{
			ID:            uuid.New(),
			EnvironmentID: envID,
			Type:          entity.ThreadTypeDirect,
		}, nil).Once()

		// Mock: Add Participants
		mockThreadRepo.On("AddParticipant", ctx, mock.MatchedBy(func(p *entity.ThreadParticipant) bool {
			return p.EntityType == "user" && p.EntityID == memberID
		})).Return(nil).Once()
		mockThreadRepo.On("AddParticipant", ctx, mock.MatchedBy(func(p *entity.ThreadParticipant) bool {
			return p.EntityType == "subscriber" && p.EntityID == targetSubID
		})).Return(nil).Once()

		thread, err := service.GetOrCreateDirectThread(ctx, envID, memberID, targetSubID, "subscriber")

		assert.NoError(t, err)
		assert.NotNil(t, thread)
		assert.Equal(t, entity.ThreadTypeDirect, thread.Type)
		mockThreadRepo.AssertExpectations(t)
	})

	t.Run("returns existing thread", func(t *testing.T) {
		existingThread := &entity.Thread{
			ID:            uuid.New(),
			EnvironmentID: envID,
			Type:          entity.ThreadTypeDirect,
		}
		// Mock: Create fails (unique constraint violation)
		mockThreadRepo.On("Create", ctx, mock.AnythingOfType("*entity.Thread")).Return((*entity.Thread)(nil), assert.AnError).Once()
		// Mock: Fallback to fetch existing
		mockThreadRepo.On("GetDirectThreadBetweenEntities", ctx, "user", memberID, "subscriber", targetSubID).Return(existingThread, nil).Once()

		thread, err := service.GetOrCreateDirectThread(ctx, envID, memberID, targetSubID, "subscriber")

		assert.NoError(t, err)
		assert.Equal(t, existingThread.ID, thread.ID)
		mockThreadRepo.AssertExpectations(t)
	})
}

func TestConversationService_CreateGroupThread(t *testing.T) {
	ctx := context.Background()
	envID := uuid.New()
	agentID := uuid.New()
	subID := uuid.New()

	mockMsgRepo := new(MockMessageRepository)
	mockThreadRepo := new(MockThreadRepository)
	mockSubRepo := new(MockSubscriberRepository)
	mockLogRepo := new(MockAssignmentLogRepository)
	mockMemberRepo := new(MockMemberRepository)
	mockAppRepo := new(MockAppRepository)
	mockEnvRepo := new(MockEnvironmentRepository)

	service := impl.NewConversationService(mockMsgRepo, mockThreadRepo, mockSubRepo, mockLogRepo, mockMemberRepo, mockAppRepo, mockEnvRepo, nil)

	t.Run("creates group with mixed participants", func(t *testing.T) {
		participants := []*entity.ThreadParticipant{
			{EntityID: agentID, EntityType: "user"},
			{EntityID: subID, EntityType: "subscriber"},
		}

		mockThreadRepo.On("Create", ctx, mock.AnythingOfType("*entity.Thread")).Return(&entity.Thread{
			ID:            uuid.New(),
			EnvironmentID: envID,
			Type:          entity.ThreadTypeGroup,
		}, nil).Once()

		mockThreadRepo.On("AddParticipant", ctx, mock.AnythingOfType("*entity.ThreadParticipant")).Return(nil).Twice()

		thread, err := service.CreateGroupThread(ctx, envID, "Group X", participants)

		assert.NoError(t, err)
		assert.NotNil(t, thread)
		assert.Equal(t, entity.ThreadTypeGroup, thread.Type)
		assert.Len(t, thread.Participants, 2)
		mockThreadRepo.AssertExpectations(t)
	})
}
