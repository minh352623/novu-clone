package jwt

import (
	"errors"
	"fmt"
	"time"

	"CONVERDA/global"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidClaims    = errors.New("invalid token claims")
	ErrTokenNotProvided = errors.New("token not provided")
)

// Claims represents custom JWT claims
type Claims struct {
	UserId string `json:"userId"` // Changed to string for UUID
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateAccessToken generates a new access token
func GenerateAccessToken(userId string, email string) (string, int64, error) {
	expiresIn, err := time.ParseDuration(global.Config.Auth.JwtExpiresIn)
	if err != nil {
		// Fallback to default 15 minutes if parsing fails
		expiresIn = 15 * time.Minute
	}
	expiresAt := time.Now().Add(expiresIn)

	claims := &Claims{
		UserId: userId,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "converda-service",
			Subject:   userId,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(global.Config.Auth.JwtSecret))
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate access token: %w", err)
	}

	return tokenString, int64(expiresIn.Seconds()), nil
}

// GenerateRefreshToken generates a new refresh token
func GenerateRefreshToken(userId string, email string) (string, time.Time, error) {
	expiresIn, err := time.ParseDuration(global.Config.Auth.JwtRefreshExpiresIn)
	if err != nil {
		// Fallback to default 24 hours if parsing fails
		expiresIn = 24 * time.Hour
	}
	expiresAt := time.Now().Add(expiresIn)

	claims := &Claims{
		UserId: userId,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "converda-service",
			Subject:   userId,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(global.Config.Auth.JwtSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrTokenNotProvided
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(global.Config.Auth.JwtSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) string {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}
