package repository

import (
	"context"
)

// WorkflowTxRepository provides access to transaction-specific repositories
type WorkflowTxRepository interface {
	Executions() ExecutionRepository
	// Add other repositories as needed within a transaction
}

// WorkflowUnitOfWork defines the interface for atomic operations in the Workflow module
type WorkflowUnitOfWork interface {
	Execute(ctx context.Context, fn func(repo WorkflowTxRepository) error) error
}
