package middleware

import (
	"net/http"
	"strings"

	"CONVERDA/pkg/jwt"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	// ContextKeyUserID is the key for user ID in context
	ContextKeyUserID = "userId"
	// ContextKeyUsername is the key for username in context
	ContextKeyUsername = "username"
	// ContextKeyEmail is the key for email in context
	ContextKeyEmail = "email"
)

// AuthMiddleware is the JWT authentication middleware
func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.NewAPIError(
				http.StatusUnauthorized,
				"Unauthorized",
				"Authorization header is required",
			))
			return
		}

		// Extract token from header
		tokenString := jwt.ExtractTokenFromHeader(authHeader)
		if tokenString == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.NewAPIError(
				http.StatusUnauthorized,
				"Unauthorized",
				"Invalid authorization header format. Use: Bearer <token>",
			))
			return
		}

		// Validate token
		claims, err := jwt.ValidateToken(tokenString)
		if err != nil {
			var message string
			switch err {
			case jwt.ErrExpiredToken:
				message = "Token has expired"
			case jwt.ErrInvalidToken:
				message = "Invalid token"
			case jwt.ErrInvalidClaims:
				message = "Invalid token claims"
			default:
				message = err.Error()
			}

			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.NewAPIError(
				http.StatusUnauthorized,
				"Unauthorized",
				message,
			))
			return
		}

		// Set user info in context
		ctx.Set(ContextKeyUserID, claims.UserId)
		ctx.Set(ContextKeyUsername, claims.Username)
		ctx.Set(ContextKeyEmail, claims.Email)

		ctx.Next()
	}
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx *gin.Context) (int64, bool) {
	userId, exists := ctx.Get(ContextKeyUserID)
	if !exists {
		return 0, false
	}
	id, ok := userId.(int64)
	return id, ok
}

// GetUsernameFromContext extracts username from context
func GetUsernameFromContext(ctx *gin.Context) (string, bool) {
	username, exists := ctx.Get(ContextKeyUsername)
	if !exists {
		return "", false
	}
	name, ok := username.(string)
	return name, ok
}

// GetEmailFromContext extracts email from context
func GetEmailFromContext(ctx *gin.Context) (string, bool) {
	email, exists := ctx.Get(ContextKeyEmail)
	if !exists {
		return "", false
	}
	e, ok := email.(string)
	return e, ok
}

// OptionalAuthMiddleware is an optional authentication middleware
// It will set user info if token is provided, but won't block if not
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.Next()
			return
		}

		tokenString := jwt.ExtractTokenFromHeader(authHeader)
		if tokenString == "" {
			ctx.Next()
			return
		}

		claims, err := jwt.ValidateToken(tokenString)
		if err != nil {
			ctx.Next()
			return
		}

		// Set user info in context
		ctx.Set(ContextKeyUserID, claims.UserId)
		ctx.Set(ContextKeyUsername, claims.Username)
		ctx.Set(ContextKeyEmail, claims.Email)

		ctx.Next()
	}
}

// ValidateAccessToken validates an access token string and returns claims
func ValidateAccessToken(tokenString string) (*jwt.Claims, error) {
	// Remove Bearer prefix if present
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = tokenString[7:]
	}
	return jwt.ValidateToken(tokenString)
}
