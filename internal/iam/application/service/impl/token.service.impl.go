package impl

import (
	"context"
	"time"

	"CONVERDA/internal/iam/application/service"
	jwtPkg "CONVERDA/pkg/jwt"

	"github.com/google/uuid"
)

type tokenServiceImpl struct {
}

func NewTokenService() service.TokenService {
	return &tokenServiceImpl{}
}

func (s *tokenServiceImpl) GenerateAccessToken(ctx context.Context, userID uuid.UUID, email string) (string, int64, error) {
	return jwtPkg.GenerateAccessToken(userID.String(), email)
}

func (s *tokenServiceImpl) GenerateRefreshToken(ctx context.Context, userID uuid.UUID, email string) (string, time.Time, error) {
	return jwtPkg.GenerateRefreshToken(userID.String(), email)
}

func (s *tokenServiceImpl) ValidateRefreshToken(ctx context.Context, tokenString string) (uuid.UUID, string, error) {
	claims, err := jwtPkg.ValidateToken(tokenString)
	if err != nil {
		return uuid.Nil, "", err
	}

	userID, err := uuid.Parse(claims.UserId)
	if err != nil {
		return uuid.Nil, "", err
	}

	return userID, claims.Email, nil
}
