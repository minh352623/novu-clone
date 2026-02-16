package service

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TokenService handles token generation and validation features
type TokenService interface {
	// GenerateAccessToken generates a new access token for a user
	GenerateAccessToken(ctx context.Context, userID uuid.UUID, email string) (string, int64, error)

	// GenerateRefreshToken generates a new refresh token for a user
	GenerateRefreshToken(ctx context.Context, userID uuid.UUID, email string) (string, time.Time, error)

	// ValidateRefreshToken validates a refresh token and returns the user ID and email
	ValidateRefreshToken(ctx context.Context, tokenString string) (uuid.UUID, string, error)
}
