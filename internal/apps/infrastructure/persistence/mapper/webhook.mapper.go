package mapper

import (
	"encoding/json"

	"CONVERDA/internal/apps/domain/model/entity"
	"CONVERDA/internal/apps/infrastructure/persistence/model"

	"gorm.io/datatypes"
)

func ToWebhookEntity(m *model.WebhookModel) *entity.Webhook {
	if m == nil {
		return nil
	}
	var events []string
	_ = json.Unmarshal(m.Events, &events)

	return &entity.Webhook{
		ID:            m.ID,
		TenantID:      m.TenantID,
		AppID:         m.AppID,
		EnvironmentID: m.EnvironmentID,
		URL:           m.URL,
		Secret:        m.Secret,
		Description:   m.Description,
		Events:        events,
		IsActive:      m.IsActive,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func ToWebhookModel(e *entity.Webhook) *model.WebhookModel {
	if e == nil {
		return nil
	}
	eventsJSON, _ := json.Marshal(e.Events)

	return &model.WebhookModel{
		ID:            e.ID,
		TenantID:      e.TenantID,
		AppID:         e.AppID,
		EnvironmentID: e.EnvironmentID,
		URL:           e.URL,
		Secret:        e.Secret,
		Description:   e.Description,
		Events:        datatypes.JSON(eventsJSON),
		IsActive:      e.IsActive,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}

func ToWebhookEntities(models []model.WebhookModel) []*entity.Webhook {
	entities := make([]*entity.Webhook, 0, len(models))
	for i := range models {
		entities = append(entities, ToWebhookEntity(&models[i]))
	}
	return entities
}
