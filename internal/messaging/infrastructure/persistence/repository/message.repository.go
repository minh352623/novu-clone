package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"CONVERDA/internal/messaging/domain/model/entity"
	domainRepo "CONVERDA/internal/messaging/domain/repository"
	"CONVERDA/internal/messaging/infrastructure/persistence/model"
	"CONVERDA/pkg/cursor"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) domainRepo.MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, msg *entity.Message) (*entity.Message, error) {
	msgModel := &model.MessageModel{
		ID:            msg.ID,
		TenantID:      msg.TenantID,
		EnvironmentID: msg.EnvironmentID,
		ThreadID:      msg.ThreadID,
		SenderType:    msg.SenderType,
		SenderID:      msg.SenderID,
		Content:       datatypes.JSON(msg.Content),
		Type:          msg.Type,
		ParentID:      msg.ParentID,
		CreatedAt:     msg.CreatedAt,
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Insert Message
		if err := tx.Create(msgModel).Error; err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}

		// 2. Insert Closure Paths
		// Self-reference (Distance 0)
		selfRef := model.MessageClosureModel{
			AncestorID:   msgModel.ID,
			DescendantID: msgModel.ID,
			Depth:        0,
		}
		if err := tx.Create(&selfRef).Error; err != nil {
			return fmt.Errorf("failed to create self-reference closure: %w", err)
		}

		// Ancestor paths (Distance + 1)
		if msg.ParentID != nil {
			// SQL:
			// INSERT INTO message_closure (ancestor_id, descendant_id, depth)
			// SELECT ancestor_id, 'NEW_ID', depth + 1
			// FROM message_closure
			// WHERE descendant_id = 'PARENT_ID'
			query := `
				INSERT INTO message_closure (ancestor_id, descendant_id, depth)
				SELECT ancestor_id, ?, depth + 1
				FROM message_closure
				WHERE descendant_id = ?
			`
			if err := tx.Exec(query, msgModel.ID, *msg.ParentID).Error; err != nil {
				return fmt.Errorf("failed to insert closure paths: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (r *messageRepository) GetMessagesByThread(ctx context.Context, threadID uuid.UUID) ([]*entity.Message, error) {
	var models []*model.MessageModel
	err := r.db.WithContext(ctx).
		Where("thread_id = ?", threadID).
		Order("created_at ASC").
		Find(&models).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get messages by thread: %w", err)
	}

	entities := make([]*entity.Message, 0, len(models))
	for _, m := range models {
		content, _ := m.Content.MarshalJSON()
		entities = append(entities, &entity.Message{
			ID:            m.ID,
			TenantID:      m.TenantID,
			EnvironmentID: m.EnvironmentID,
			ThreadID:      m.ThreadID,
			SenderType:    m.SenderType,
			SenderID:      m.SenderID,
			Content:       json.RawMessage(content),
			ParentID:      m.ParentID,
			CreatedAt:     m.CreatedAt,
		})
	}
	return entities, nil
}

func (r *messageRepository) ListByThread(ctx context.Context, threadID uuid.UUID, limit, offset int) ([]*entity.Message, error) {
	var models []*model.MessageModel
	err := r.db.WithContext(ctx).
		Where("thread_id = ?", threadID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list messages by thread: %w", err)
	}

	entities := make([]*entity.Message, 0, len(models))
	for _, m := range models {
		content, _ := m.Content.MarshalJSON()
		entities = append(entities, &entity.Message{
			ID:            m.ID,
			TenantID:      m.TenantID,
			EnvironmentID: m.EnvironmentID,
			ThreadID:      m.ThreadID,
			SenderType:    m.SenderType,
			SenderID:      m.SenderID,
			Content:       json.RawMessage(content),
			ParentID:      m.ParentID,
			CreatedAt:     m.CreatedAt,
		})
	}
	return entities, nil
}

func (r *messageRepository) ListByCursor(ctx context.Context, threadID uuid.UUID, c *cursor.Cursor, direction string, limit int) ([]*entity.Message, error) {
	var models []*model.MessageModel
	query := r.db.WithContext(ctx).Where("thread_id = ?", threadID)

	if c != nil {
		createdAt := time.UnixMicro(c.Timestamp)
		if direction == "after" {
			// Newer messages (scrolling down / load updates)
			// (created_at > T OR (created_at = T AND id > ID))
			query = query.Where("(created_at > ? OR (created_at = ? AND id > ?))", createdAt, createdAt, c.ID).
				Order("created_at ASC, id ASC")
		} else {
			// Older messages (scrolling up / load history)
			// (created_at < T OR (created_at = T AND id < ID))
			query = query.Where("(created_at < ? OR (created_at = ? AND id < ?))", createdAt, createdAt, c.ID).
				Order("created_at DESC, id DESC")
		}
	} else {
		// Default (no cursor): Latest messages first
		query = query.Order("created_at DESC, id DESC")
	}

	if err := query.Limit(limit).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list messages by cursor: %w", err)
	}

	entities := make([]*entity.Message, 0, len(models))
	for _, m := range models {
		content, _ := m.Content.MarshalJSON()
		entities = append(entities, &entity.Message{
			ID:            m.ID,
			TenantID:      m.TenantID,
			EnvironmentID: m.EnvironmentID,
			ThreadID:      m.ThreadID,
			SenderType:    m.SenderType,
			SenderID:      m.SenderID,
			Content:       json.RawMessage(content),
			ParentID:      m.ParentID,
			CreatedAt:     m.CreatedAt,
		})
	}
	return entities, nil
}
