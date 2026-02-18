package repository

import (
	"context"
)

// AppsUnitOfWork defines the interface for atomic transactions in the Apps module.
// It follows the Unit of Work pattern to ensure ACID properties for multi-step operations.
type AppsUnitOfWork interface {
	// Execute runs the given function within a transaction.
	// If the function returns an error, the transaction is rolled back.
	// Otherwise, it is committed.
	Execute(ctx context.Context, fn func(repo AppsTxRepository) error) error
}

// AppsTxRepository provides access to all Apps repositories within a single transaction context.
type AppsTxRepository interface {
	Apps() AppRepository
	Environments() EnvironmentRepository
	APIKeys() APIKeyRepository
	Webhooks() WebhookRepository
	Providers() ProviderRepository
	Metrics() UsageMetricRepository
}
