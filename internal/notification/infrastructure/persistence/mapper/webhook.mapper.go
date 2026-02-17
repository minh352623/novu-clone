package mapper

import (
	"encoding/json"

	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/infrastructure/persistence/model"

	"gorm.io/datatypes"
)

func ToWebhookDomain(m *model.WebhookModel) *entity.Webhook {
	if m == nil {
		return nil
	}
	var events []string
	if len(m.Events) > 0 {
		_ = json.Unmarshal(m.Events, &events)
	}

	return &entity.Webhook{
		ID:                  m.ID,
		TenantID:            m.TenantID,
		EnvironmentID:       m.EnvironmentID,
		AppID:               m.AppID,
		URL:                 m.URL,
		Secret:              m.Secret,
		Events:              events,
		Description:         m.Description,
		IsActive:            m.IsActive,
		MaxRetries:          m.MaxRetries,
		RetryBackoffSeconds: m.RetryBackoffSeconds,
		RetryTimeoutSeconds: m.RetryTimeoutSeconds,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}

func ToWebhookLogDomain(m *model.WebhookLogModel) *entity.WebhookLog {
	if m == nil {
		return nil
	}
	var payload map[string]interface{}
	if len(m.RequestPayload) > 0 {
		_ = json.Unmarshal(m.RequestPayload, &payload)
	}

	return &entity.WebhookLog{
		ID:             m.ID,
		WebhookID:      m.WebhookID,
		URL:            m.URL,
		EventType:      m.EventType,
		RequestPayload: payload,
		ResponseCode:   m.ResponseCode,
		ResponseBody:   m.ResponseBody,
		DurationMs:     m.DurationMs,
		Status:         m.Status,
		RetryCount:     m.RetryCount,
		MaxRetries:     m.MaxRetries,
		NextRetryAt:    m.NextRetryAt,
		TenantID:       m.TenantID,
		AppID:          m.AppID,
		CreatedAt:      m.CreatedAt,
	}
}

func ToWebhookLogModel(e *entity.WebhookLog) *model.WebhookLogModel {
	if e == nil {
		return nil
	}
	var payloadJSON datatypes.JSON
	if e.RequestPayload != nil {
		bytes, _ := json.Marshal(e.RequestPayload)
		payloadJSON = datatypes.JSON(bytes)
	}

	return &model.WebhookLogModel{
		ID:             e.ID,
		WebhookID:      e.WebhookID,
		URL:            e.URL,
		EventType:      e.EventType,
		RequestPayload: payloadJSON,
		ResponseCode:   e.ResponseCode,
		ResponseBody:   e.ResponseBody,
		DurationMs:     e.DurationMs,
		Status:         e.Status,
		RetryCount:     e.RetryCount,
		MaxRetries:     e.MaxRetries,
		NextRetryAt:    e.NextRetryAt,
		TenantID:       e.TenantID,
		AppID:          e.AppID,
		CreatedAt:      e.CreatedAt,
	}
}
