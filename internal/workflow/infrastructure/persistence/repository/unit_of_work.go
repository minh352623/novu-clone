package repository

import (
	"context"

	"CONVERDA/internal/workflow/domain/repository"

	"gorm.io/gorm"
)

type workflowTxRepository struct {
	tx *gorm.DB
}

func (r *workflowTxRepository) Executions() repository.ExecutionRepository {
	return NewExecutionRepository(r.tx)
}

type workflowUnitOfWork struct {
	db *gorm.DB
}

func NewWorkflowUnitOfWork(db *gorm.DB) repository.WorkflowUnitOfWork {
	return &workflowUnitOfWork{db: db}
}

func (uow *workflowUnitOfWork) Execute(ctx context.Context, fn func(repo repository.WorkflowTxRepository) error) error {
	return uow.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &workflowTxRepository{tx: tx}
		if err := fn(txRepo); err != nil {
			return err
		}
		return nil
	})
}
