package impl_test

import (
	"context"
	"os"
	"testing"
	"time"

	"CONVERDA/global"
	"CONVERDA/internal/workflow/application/service/impl"
	"CONVERDA/internal/workflow/domain/model/entity"
	domainRepo "CONVERDA/internal/workflow/domain/repository"
	"CONVERDA/pkg/logger"
	"CONVERDA/pkg/setting"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMain(m *testing.M) {
	global.Logger = logger.NewLogger(setting.LoggerSetting{
		LogLevel: "debug",
	})
	os.Exit(m.Run())
}

// === Mocks ===

type mockWorkflowRepo struct {
	mock.Mock
}

func (m *mockWorkflowRepo) Create(ctx context.Context, w *entity.Workflow) error {
	return m.Called(ctx, w).Error(0)
}
func (m *mockWorkflowRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Workflow, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Workflow), args.Error(1)
}
func (m *mockWorkflowRepo) ListByEnvironment(ctx context.Context, envID uuid.UUID) ([]*entity.Workflow, error) {
	args := m.Called(ctx, envID)
	return args.Get(0).([]*entity.Workflow), args.Error(1)
}
func (m *mockWorkflowRepo) Update(ctx context.Context, w *entity.Workflow) error {
	return m.Called(ctx, w).Error(0)
}
func (m *mockWorkflowRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockWorkflowRepo) GetByTrigger(ctx context.Context, envID uuid.UUID, trigger string) (*entity.Workflow, error) {
	args := m.Called(ctx, envID, trigger)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Workflow), args.Error(1)
}
func (m *mockWorkflowRepo) AddStep(ctx context.Context, s *entity.WorkflowStep) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockWorkflowRepo) UpdateStep(ctx context.Context, s *entity.WorkflowStep) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockWorkflowRepo) DeleteStep(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockWorkflowRepo) GetStepsByWorkflow(ctx context.Context, wid uuid.UUID) ([]entity.WorkflowStep, error) {
	args := m.Called(ctx, wid)
	return args.Get(0).([]entity.WorkflowStep), args.Error(1)
}

type mockExecRepo struct {
	mock.Mock
}

func (m *mockExecRepo) CreateExecution(ctx context.Context, exec *entity.WorkflowExecution) error {
	return m.Called(ctx, exec).Error(0)
}
func (m *mockExecRepo) GetExecution(ctx context.Context, id uuid.UUID) (*entity.WorkflowExecution, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkflowExecution), args.Error(1)
}
func (m *mockExecRepo) UpdateExecution(ctx context.Context, exec *entity.WorkflowExecution) error {
	return m.Called(ctx, exec).Error(0)
}
func (m *mockExecRepo) CreateStepExecution(ctx context.Context, se *entity.StepExecution) error {
	return m.Called(ctx, se).Error(0)
}
func (m *mockExecRepo) UpdateStepExecution(ctx context.Context, se *entity.StepExecution) error {
	return m.Called(ctx, se).Error(0)
}
func (m *mockExecRepo) GetPendingScheduledSteps(ctx context.Context, limit int) ([]*entity.StepExecution, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*entity.StepExecution), args.Error(1)
}
func (m *mockExecRepo) GetStepExecutionsByExecution(ctx context.Context, execID uuid.UUID) ([]*entity.StepExecution, error) {
	args := m.Called(ctx, execID)
	return args.Get(0).([]*entity.StepExecution), args.Error(1)
}
func (m *mockExecRepo) BufferDigestEvent(ctx context.Context, event *entity.DigestEvent) error {
	return m.Called(ctx, event).Error(0)
}
func (m *mockExecRepo) FlushDigestEvents(ctx context.Context, stepID, executionID uuid.UUID) ([]*entity.DigestEvent, error) {
	args := m.Called(ctx, stepID, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.DigestEvent), args.Error(1)
}
func (m *mockExecRepo) GetDigestingSteps(ctx context.Context, limit int) ([]*entity.StepExecution, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*entity.StepExecution), args.Error(1)
}

// === UoW Mocks ===

type mockWorkflowTxRepository struct {
	mock.Mock
	MockExecRepo *mockExecRepo
}

func (m *mockWorkflowTxRepository) Executions() domainRepo.ExecutionRepository {
	return m.MockExecRepo
}

type mockWorkflowUow struct {
	mock.Mock
	TxRepo domainRepo.WorkflowTxRepository
}

func (m *mockWorkflowUow) Execute(ctx context.Context, fn func(repo domainRepo.WorkflowTxRepository) error) error {
	return fn(m.TxRepo)
}

type mockStepHandler struct {
	mock.Mock
}

func (m *mockStepHandler) Execute(ctx context.Context, step *entity.WorkflowStep,
	exec *entity.WorkflowExecution, stepExec *entity.StepExecution) error {
	return m.Called(ctx, step, exec, stepExec).Error(0)
}

// === Tests ===

func TestTriggerService_Trigger_ChannelStep(t *testing.T) {
	ctx := context.Background()
	envID := uuid.New()
	stepID := uuid.New()

	wfRepo := new(mockWorkflowRepo)
	execRepo := new(mockExecRepo)
	uow := &mockWorkflowUow{TxRepo: &mockWorkflowTxRepository{MockExecRepo: execRepo}}
	channelHandler := new(mockStepHandler)
	delayHandler := new(mockStepHandler)

	workflow := &entity.Workflow{
		ID:                uuid.New(),
		EnvironmentID:     envID,
		TriggerIdentifier: "user.signup",
		IsActive:          true,
		Steps: []entity.WorkflowStep{
			{
				ID:       stepID,
				StepType: entity.StepTypeChannel,
				Config:   map[string]interface{}{"template_code": "welcome", "channel": "email"},
				Order:    0,
			},
		},
	}

	wfRepo.On("GetByTrigger", ctx, envID, "user.signup").Return(workflow, nil)
	execRepo.On("CreateExecution", ctx, mock.Anything).Return(nil)
	execRepo.On("CreateStepExecution", ctx, mock.Anything).Return(nil)
	execRepo.On("UpdateExecution", ctx, mock.Anything).Return(nil)
	execRepo.On("UpdateStepExecution", ctx, mock.Anything).Return(nil)
	channelHandler.On("Execute", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := impl.NewTriggerService(wfRepo, execRepo, uow, channelHandler, delayHandler)
	err := svc.Trigger(ctx, envID, "user.signup", "user-123", map[string]interface{}{"name": "John"})

	assert.NoError(t, err)
	wfRepo.AssertCalled(t, "GetByTrigger", ctx, envID, "user.signup")
	channelHandler.AssertCalled(t, "Execute", ctx, mock.Anything, mock.Anything, mock.Anything)
}

func TestDelayHandler_ParsesDuration(t *testing.T) {
	ctx := context.Background()
	execRepo := new(mockExecRepo)
	handler := impl.NewDelayHandler(execRepo)

	stepExec := &entity.StepExecution{
		ID:     uuid.New(),
		Status: entity.StepStatusRunning,
	}

	step := &entity.WorkflowStep{
		ID:       uuid.New(),
		StepType: entity.StepTypeDelay,
		Config:   map[string]interface{}{"duration": "1h"},
	}

	exec := &entity.WorkflowExecution{
		ID:     uuid.New(),
		Status: entity.WorkflowExecutionRunning,
	}

	execRepo.On("UpdateStepExecution", ctx, mock.Anything).Return(nil)

	err := handler.Execute(ctx, step, exec, stepExec)

	assert.NoError(t, err)
	assert.Equal(t, entity.StepStatusScheduled, stepExec.Status)
	assert.NotNil(t, stepExec.ScheduledAt)
	// Should be ~1 hour from now
	assert.WithinDuration(t, time.Now().Add(1*time.Hour), *stepExec.ScheduledAt, 5*time.Second)
}

func TestDelayHandler_ParsesDurationSeconds(t *testing.T) {
	ctx := context.Background()
	execRepo := new(mockExecRepo)
	handler := impl.NewDelayHandler(execRepo)

	stepExec := &entity.StepExecution{
		ID:     uuid.New(),
		Status: entity.StepStatusRunning,
	}

	step := &entity.WorkflowStep{
		ID:       uuid.New(),
		StepType: entity.StepTypeDelay,
		Config:   map[string]interface{}{"duration_seconds": float64(3600)},
	}

	exec := &entity.WorkflowExecution{ID: uuid.New()}
	execRepo.On("UpdateStepExecution", ctx, mock.Anything).Return(nil)

	err := handler.Execute(ctx, step, exec, stepExec)

	assert.NoError(t, err)
	assert.Equal(t, entity.StepStatusScheduled, stepExec.Status)
	assert.WithinDuration(t, time.Now().Add(1*time.Hour), *stepExec.ScheduledAt, 5*time.Second)
}

func TestDelayHandler_MissingDuration(t *testing.T) {
	ctx := context.Background()
	execRepo := new(mockExecRepo)
	handler := impl.NewDelayHandler(execRepo)

	stepExec := &entity.StepExecution{ID: uuid.New()}
	step := &entity.WorkflowStep{
		ID:     uuid.New(),
		Config: map[string]interface{}{},
	}
	exec := &entity.WorkflowExecution{ID: uuid.New()}

	err := handler.Execute(ctx, step, exec, stepExec)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing duration")
}
