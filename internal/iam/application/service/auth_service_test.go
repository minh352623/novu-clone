package service_test

import (
	"context"
	"errors"
	"testing"

	"CONVERDA/internal/iam/application/service"
	"CONVERDA/internal/iam/application/service/impl"
	"CONVERDA/internal/iam/domain/model/entity"
	"CONVERDA/internal/iam/domain/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetAll(ctx context.Context, filters repository.UserFilters) ([]*entity.User, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*entity.User), args.Error(1)
}

func (m *MockUserRepository) CountAll(ctx context.Context, filters repository.UserFilters) (int64, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	args := m.Called(ctx, id, passwordHash)
	return args.Error(0)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		password      string
		fullName      *string
		setupMock     func(*MockUserRepository)
		expectedError error
		expectUser    bool
	}{
		{
			name:     "successful registration",
			email:    "test@example.com",
			password: "password123",
			fullName: nil,
			setupMock: func(m *MockUserRepository) {
				m.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, nil)
				m.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(&entity.User{
					ID:    uuid.New(),
					Email: "test@example.com",
				}, nil)
			},
			expectedError: nil,
			expectUser:    true,
		},
		{
			name:     "user already exists",
			email:    "existing@example.com",
			password: "password123",
			fullName: nil,
			setupMock: func(m *MockUserRepository) {
				m.On("ExistsByEmail", mock.Anything, "existing@example.com").Return(true, nil)
			},
			expectedError: service.ErrUserAlreadyExists,
			expectUser:    false,
		},
		{
			name:     "password too short",
			email:    "test@example.com",
			password: "short",
			fullName: nil,
			setupMock: func(m *MockUserRepository) {
				m.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, nil)
			},
			expectedError: entity.ErrUserPasswordTooShort,
			expectUser:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)

			authService := impl.NewAuthService(mockRepo)
			user, err := authService.Register(context.Background(), tt.email, tt.password, tt.fullName)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError) || err.Error() == tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.expectUser {
				assert.NotNil(t, user)
			} else {
				assert.Nil(t, user)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	// Create a valid user with hashed password
	validPassword := "password123"
	validUser, _ := entity.NewUser("test@example.com", validPassword, nil)

	tests := []struct {
		name          string
		email         string
		password      string
		setupMock     func(*MockUserRepository)
		expectedError error
		expectUser    bool
	}{
		{
			name:     "successful login",
			email:    "test@example.com",
			password: validPassword,
			setupMock: func(m *MockUserRepository) {
				m.On("GetByEmail", mock.Anything, "test@example.com").Return(validUser, nil)
				m.On("UpdateLastLogin", mock.Anything, validUser.ID).Return(nil)
			},
			expectedError: nil,
			expectUser:    true,
		},
		{
			name:     "user not found",
			email:    "notfound@example.com",
			password: "password123",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, nil)
			},
			expectedError: service.ErrInvalidCredentials,
			expectUser:    false,
		},
		{
			name:     "wrong password",
			email:    "test@example.com",
			password: "wrongpassword",
			setupMock: func(m *MockUserRepository) {
				m.On("GetByEmail", mock.Anything, "test@example.com").Return(validUser, nil)
			},
			expectedError: service.ErrInvalidCredentials,
			expectUser:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)

			authService := impl.NewAuthService(mockRepo)
			user, err := authService.Login(context.Background(), tt.email, tt.password)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectUser {
				assert.NotNil(t, user)
			} else {
				assert.Nil(t, user)
			}
		})
	}
}
