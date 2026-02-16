package repository

import (
	"context"

	"CONVERDA/internal/messaging/domain/model/entity"
	"CONVERDA/pkg/cursor"

	"github.com/google/uuid"
)

type MessageRepository interface {
	Create(ctx context.Context, msg *entity.Message) (*entity.Message, error)
	GetMessagesByThread(ctx context.Context, threadID uuid.UUID) ([]*entity.Message, error)
	ListByThread(ctx context.Context, threadID uuid.UUID, limit, offset int) ([]*entity.Message, error)
	ListByCursor(ctx context.Context, threadID uuid.UUID, cursor *cursor.Cursor, direction string, limit int) ([]*entity.Message, error)
}
