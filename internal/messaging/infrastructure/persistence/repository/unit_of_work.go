package repository

import (
	"context"

	"CONVERDA/internal/messaging/domain/repository"

	"gorm.io/gorm"
)

type gormMessagingUnitOfWork struct {
	db *gorm.DB
}

// NewMessagingUnitOfWork creates a new GORM-based implementation of MessagingUnitOfWork
func NewMessagingUnitOfWork(db *gorm.DB) repository.MessagingUnitOfWork {
	return &gormMessagingUnitOfWork{db: db}
}

func (u *gormMessagingUnitOfWork) Execute(ctx context.Context, fn func(repository.MessagingTxRepository) error) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &gormMessagingTxRepository{
			messages:       NewMessageRepository(tx),
			threads:        NewThreadRepository(tx),
			subscribers:    NewSubscriberRepository(tx),
			assignmentLogs: NewAssignmentLogRepository(tx),
		}
		return fn(txRepo)
	})
}

type gormMessagingTxRepository struct {
	messages       repository.MessageRepository
	threads        repository.ThreadRepository
	subscribers    repository.SubscriberRepository
	assignmentLogs repository.AssignmentLogRepository
}

func (r *gormMessagingTxRepository) Messages() repository.MessageRepository {
	return r.messages
}

func (r *gormMessagingTxRepository) Threads() repository.ThreadRepository {
	return r.threads
}

func (r *gormMessagingTxRepository) Subscribers() repository.SubscriberRepository {
	return r.subscribers
}

func (r *gormMessagingTxRepository) AssignmentLogs() repository.AssignmentLogRepository {
	return r.assignmentLogs
}
