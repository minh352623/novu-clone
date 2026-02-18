package repository

import (
	"context"
)

// MessagingUnitOfWork defines the interface for atomic transactions in the Messaging module.
// It follows the Unit of Work pattern to ensure ACID properties for multi-step operations.
type MessagingUnitOfWork interface {
	// Execute runs the given function within a transaction.
	// If the function returns an error, the transaction is rolled back.
	// Otherwise, it is committed.
	Execute(ctx context.Context, fn func(repo MessagingTxRepository) error) error
}

// MessagingTxRepository provides access to all Messaging repositories within a single transaction context.
type MessagingTxRepository interface {
	Messages() MessageRepository
	Threads() ThreadRepository
	Subscribers() SubscriberRepository
	AssignmentLogs() AssignmentLogRepository
}
