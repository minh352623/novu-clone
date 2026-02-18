package repository

import (
	"context"
)

// IAMUnitOfWork defines the interface for atomic transactions in the IAM module.
// It follows the Unit of Work pattern to ensure ACID properties for multi-step operations.
type IAMUnitOfWork interface {
	// Execute runs the given function within a transaction.
	// If the function returns an error, the transaction is rolled back.
	// Otherwise, it is committed.
	Execute(ctx context.Context, fn func(repo IAMTxRepository) error) error
}

// IAMTxRepository provides access to all IAM repositories within a single transaction context.
type IAMTxRepository interface {
	Members() TenantMemberRepository
	Invitations() InvitationRepository
	Users() UserRepository
	Roles() RoleRepository
	Tenants() TenantRepository
}
