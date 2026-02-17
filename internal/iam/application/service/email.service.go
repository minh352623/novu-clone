package service

import (
	"context"
)

// EmailService handles email operations
type EmailService interface {
	// SendInvitation sends an invitation email to a user
	SendInvitation(ctx context.Context, toEmail, token, roleName, tenantName, inviterName string) error
}
