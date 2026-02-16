package mapper

import (
	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/infrastructure/persistence/model"
)

// NotificationMapper
type NotificationMapper struct{}

func NewNotificationMapper() *NotificationMapper {
	return &NotificationMapper{}
}

func (m *NotificationMapper) ToDomain(l *model.NotificationModel) *entity.Notification {
	return &entity.Notification{
		ID:            l.ID,
		TenantID:      l.TenantID,
		EnvironmentID: l.EnvironmentID,
		TemplateCode:  l.TemplateCode,
		Recipient:     l.Recipient,
		Channel:       l.Channel,
		Status:        l.Status,
		Data:          l.Data,
		ErrorMessage:  l.ErrorMessage,
		SentAt:        l.SentAt,
		CreatedAt:     l.CreatedAt,
		UpdatedAt:     l.UpdatedAt,
	}
}

func (m *NotificationMapper) ToModel(d *entity.Notification) *model.NotificationModel {
	return &model.NotificationModel{
		ID:            d.ID,
		TenantID:      d.TenantID,
		EnvironmentID: d.EnvironmentID,
		TemplateCode:  d.TemplateCode,
		Recipient:     d.Recipient,
		Channel:       d.Channel,
		Status:        d.Status,
		Data:          d.Data,
		ErrorMessage:  d.ErrorMessage,
		SentAt:        d.SentAt,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}
