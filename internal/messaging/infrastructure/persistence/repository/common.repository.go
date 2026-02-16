package repository

import (
	"context"
	"encoding/json"
	"time"

	"CONVERDA/internal/messaging/domain/model/entity"
	domainRepo "CONVERDA/internal/messaging/domain/repository"
	"CONVERDA/internal/messaging/infrastructure/persistence/model"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// --- SUBSCRIBER REPO IMPL ---

type subscriberRepository struct {
	db *gorm.DB
}

func NewSubscriberRepository(db *gorm.DB) domainRepo.SubscriberRepository {
	return &subscriberRepository{db: db}
}

func (r *subscriberRepository) Create(ctx context.Context, sub *entity.Subscriber) (*entity.Subscriber, error) {
	m := &model.SubscriberModel{
		ID:            sub.ID,
		EnvironmentID: sub.EnvironmentID,
		SubscriberKey: sub.SubscriberKey,
		Email:         sub.Email,
		Phone:         sub.Phone,
		Data:          datatypes.JSON(sub.Data),
		CreatedAt:     sub.CreatedAt,
		UpdatedAt:     sub.UpdatedAt,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return sub, nil
}

func (r *subscriberRepository) GetByKey(ctx context.Context, envID uuid.UUID, key string) (*entity.Subscriber, error) {
	var m model.SubscriberModel
	if err := r.db.WithContext(ctx).Where("environment_id = ? AND subscriber_key = ?", envID, key).First(&m).Error; err != nil {
		return nil, err
	}
	data, _ := m.Data.MarshalJSON()
	return &entity.Subscriber{
		ID:            m.ID,
		EnvironmentID: m.EnvironmentID,
		SubscriberKey: m.SubscriberKey,
		Email:         m.Email,
		Phone:         m.Phone,
		Data:          data,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}, nil
}

func (r *subscriberRepository) Update(ctx context.Context, sub *entity.Subscriber) error {
	m := &model.SubscriberModel{
		ID:            sub.ID,
		EnvironmentID: sub.EnvironmentID,
		SubscriberKey: sub.SubscriberKey,
		Email:         sub.Email,
		Phone:         sub.Phone,
		Data:          datatypes.JSON(sub.Data),
		UpdatedAt:     time.Now(),
	}
	return r.db.WithContext(ctx).Model(m).Updates(m).Error
}

// --- THREAD REPO IMPL ---

type threadRepository struct {
	db *gorm.DB
}

func NewThreadRepository(db *gorm.DB) domainRepo.ThreadRepository {
	return &threadRepository{db: db}
}

func (r *threadRepository) Create(ctx context.Context, thread *entity.Thread) (*entity.Thread, error) {
	m := &model.ThreadModel{
		ID:            thread.ID,
		EnvironmentID: thread.EnvironmentID,
		Type:          thread.Type,
		Status:        thread.Status,
		Metadata:      datatypes.JSON(thread.Metadata),
		ReferenceHash: thread.ReferenceHash,
		CreatedAt:     thread.CreatedAt,
		UpdatedAt:     thread.UpdatedAt,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return thread, nil
}

func (r *threadRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Thread, error) {
	var m model.ThreadModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &entity.Thread{
		ID:            m.ID,
		EnvironmentID: m.EnvironmentID,
		Type:          m.Type,
		Status:        m.Status,
		Metadata:      json.RawMessage(m.Metadata),
		ReferenceHash: m.ReferenceHash,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}, nil
}

func (r *threadRepository) GetByIDAndEnv(ctx context.Context, id, envID uuid.UUID) (*entity.Thread, error) {
	var m model.ThreadModel
	if err := r.db.WithContext(ctx).Where("id = ? AND environment_id = ?", id, envID).First(&m).Error; err != nil {
		return nil, err
	}
	return &entity.Thread{
		ID:            m.ID,
		EnvironmentID: m.EnvironmentID,
		Type:          m.Type,
		Status:        m.Status,
		Metadata:      json.RawMessage(m.Metadata),
		ReferenceHash: m.ReferenceHash,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}, nil
}

func (r *threadRepository) Update(ctx context.Context, thread *entity.Thread) error {
	m := &model.ThreadModel{
		ID:            thread.ID,
		EnvironmentID: thread.EnvironmentID,
		Type:          thread.Type,
		Status:        thread.Status,
		Metadata:      datatypes.JSON(thread.Metadata),
		UpdatedAt:     time.Now(),
	}
	return r.db.WithContext(ctx).Model(m).Updates(m).Error
}

func (r *threadRepository) List(ctx context.Context, filter domainRepo.ThreadFilter) ([]*entity.Thread, int64, error) {
	db := r.db.WithContext(ctx).Model(&model.ThreadModel{})

	if filter.EnvironmentID != nil {
		db = db.Where("environment_id = ?", *filter.EnvironmentID)
	}
	if filter.Status != nil {
		db = db.Where("status = ?", *filter.Status)
	}
	if filter.Type != nil {
		db = db.Where("type = ?", *filter.Type)
	}

	if filter.AssignedToMe && filter.MemberID != nil {
		db = db.Where("id IN (SELECT thread_id FROM thread_participants WHERE entity_type = 'user' AND entity_id = ?)", *filter.MemberID)
	}

	var total int64
	db.Count(&total)

	var models []model.ThreadModel
	if err := db.Limit(filter.Limit).Offset(filter.Offset).Order("updated_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	var threads []*entity.Thread
	for _, m := range models {
		threads = append(threads, &entity.Thread{
			ID:            m.ID,
			EnvironmentID: m.EnvironmentID,
			Type:          m.Type,
			Status:        m.Status,
			Metadata:      json.RawMessage(m.Metadata),
			CreatedAt:     m.CreatedAt,
			UpdatedAt:     m.UpdatedAt,
		})
	}

	return threads, total, nil
}

func (r *threadRepository) AddParticipant(ctx context.Context, p *entity.ThreadParticipant) error {
	m := &model.ThreadParticipantModel{
		ID:         p.ID,
		ThreadID:   p.ThreadID,
		EntityType: p.EntityType,
		EntityID:   p.EntityID,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *threadRepository) UpdateParticipant(ctx context.Context, p *entity.ThreadParticipant) error {
	m := &model.ThreadParticipantModel{
		ID:         p.ID,
		ThreadID:   p.ThreadID,
		EntityType: p.EntityType,
		EntityID:   p.EntityID,
		LastReadAt: p.LastReadAt,
		UpdatedAt:  time.Now(),
	}
	return r.db.WithContext(ctx).Model(m).Updates(m).Error
}

func (r *threadRepository) RemoveParticipant(ctx context.Context, threadID uuid.UUID, entityType string, entityID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("thread_id = ? AND entity_type = ? AND entity_id = ?", threadID, entityType, entityID).Delete(&model.ThreadParticipantModel{}).Error
}

func (r *threadRepository) GetParticipants(ctx context.Context, threadID uuid.UUID) ([]*entity.ThreadParticipant, error) {
	var models []model.ThreadParticipantModel
	if err := r.db.WithContext(ctx).Where("thread_id = ?", threadID).Find(&models).Error; err != nil {
		return nil, err
	}
	var res []*entity.ThreadParticipant
	for _, m := range models {
		res = append(res, &entity.ThreadParticipant{
			ID:         m.ID,
			ThreadID:   m.ThreadID,
			EntityType: m.EntityType,
			EntityID:   m.EntityID,
			LastReadAt: m.LastReadAt,
			CreatedAt:  m.CreatedAt,
			UpdatedAt:  m.UpdatedAt,
		})
	}
	return res, nil
}

func (r *threadRepository) GetDirectThreadBetweenEntities(ctx context.Context, typeA string, idA uuid.UUID, typeB string, idB uuid.UUID) (*entity.Thread, error) {
	var m model.ThreadModel
	err := r.db.WithContext(ctx).
		Table("conversation_pools").
		Select("conversation_pools.*").
		Joins("JOIN thread_participants p1 ON conversation_pools.id = p1.thread_id").
		Joins("JOIN thread_participants p2 ON conversation_pools.id = p2.thread_id").
		Where("conversation_pools.type = ?", entity.ThreadTypeDirect).
		Where("p1.entity_type = ? AND p1.entity_id = ?", typeA, idA).
		Where("p2.entity_type = ? AND p2.entity_id = ?", typeB, idB).
		Limit(1).
		First(&m).Error

	if err != nil {
		return nil, err
	}

	return &entity.Thread{
		ID:            m.ID,
		EnvironmentID: m.EnvironmentID,
		Type:          m.Type,
		Status:        m.Status,
		Metadata:      json.RawMessage(m.Metadata),
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}, nil
}
