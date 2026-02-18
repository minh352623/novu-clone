package impl

import (
	"context"
	"fmt"

	"CONVERDA/global"
	"CONVERDA/internal/iam/application/service"
	"CONVERDA/internal/iam/domain/model/entity"
	"CONVERDA/internal/iam/domain/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// authServiceImpl implements AuthService
type authServiceImpl struct {
	userReader repository.UserReader
	userWriter repository.UserWriter
}

// NewAuthService creates a new AuthService
func NewAuthService(userRepo repository.UserRepository) service.AuthService {
	return &authServiceImpl{
		userReader: userRepo,
		userWriter: userRepo,
	}
}

func (s *authServiceImpl) Register(ctx context.Context, email, password string, fullName *string) (*entity.User, error) {
	// Check if user exists
	exists, err := s.userReader.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return nil, service.ErrUserAlreadyExists
	}

	// Create user entity (password hashing is done in entity)
	user, err := entity.NewUser(email, password, fullName)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Persist user
	createdUser, err := s.userWriter.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return createdUser, nil
}

func (s *authServiceImpl) Login(ctx context.Context, email, password string) (*entity.User, error) {
	user, err := s.userReader.GetByEmail(ctx, email)
	if err != nil {
		global.Logger.Error("auth: failed to get user by email", zap.String("email", email), zap.Error(err))
		return nil, service.ErrInvalidCredentials
	}
	if user == nil {
		return nil, service.ErrInvalidCredentials
	}

	if err := user.VerifyPassword(password); err != nil {
		return nil, service.ErrInvalidCredentials
	}

	// Record login
	user.RecordLogin()
	if err := s.userWriter.UpdateLastLogin(ctx, user.ID); err != nil {
		global.Logger.Warn("auth: failed to update last login", zap.String("userID", user.ID.String()), zap.Error(err))
	}

	return user, nil
}

func (s *authServiceImpl) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userReader.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return service.ErrUserNotFound
	}

	// Verify old password
	if err := user.VerifyPassword(oldPassword); err != nil {
		return service.ErrInvalidCredentials
	}

	// Update password
	if err := user.UpdatePassword(newPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if err := s.userWriter.UpdatePassword(ctx, userID, user.PasswordHash); err != nil {
		return fmt.Errorf("failed to save password: %w", err)
	}

	return nil
}

func (s *authServiceImpl) GetUser(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	user, err := s.userReader.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, service.ErrUserNotFound
	}
	return user, nil
}
