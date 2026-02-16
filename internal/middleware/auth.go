package middleware

import (
	"net/http"
	"strings"

	"CONVERDA/pkg/jwt"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// ContextKeyUserID is the key for user ID in context
	ContextKeyUserID = "user_id" // Consistent with what controller uses
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
		// Parse userID string to UUID format if needed, but for now store as is or parsed UUID
		// Controller expects UUID, so let's try to parse here or let controller parse
		// Reviewing controller: userID, exists := ctx.Get("user_id"); userID.(uuid.UUID)
		// So middleware MUST set uuid.UUID
		uid, err := uuid.Parse(claims.UserId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.NewAPIError(
				http.StatusUnauthorized,
				"Unauthorized",
				"Invalid user ID in token",
			))
			return
		}

		ctx.Set(ContextKeyUserID, uid)
		ctx.Set(ContextKeyEmail, claims.Email)

		ctx.Next()
	}
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx *gin.Context) (uuid.UUID, bool) {
	userId, exists := ctx.Get(ContextKeyUserID)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := userId.(uuid.UUID)
	return id, ok
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
		uid, err := uuid.Parse(claims.UserId)
		if err == nil {
			ctx.Set(ContextKeyUserID, uid)
			ctx.Set(ContextKeyEmail, claims.Email)
		}

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
