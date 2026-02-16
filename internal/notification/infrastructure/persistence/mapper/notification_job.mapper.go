package mapper

import (
	"encoding/json"

	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/infrastructure/persistence/model"

	"gorm.io/datatypes"
)

func ToJobDomain(m *model.NotificationJobModel) *entity.NotificationJob {
	if m == nil {
		return nil
	}

	var metadata map[string]interface{}
	if len(m.Metadata) > 0 {
		_ = json.Unmarshal(m.Metadata, &metadata)
	}

	var recipientsData map[string]interface{}
	if len(m.RecipientsData) > 0 {
		_ = json.Unmarshal(m.RecipientsData, &recipientsData)
	}

	return &entity.NotificationJob{
		ID:             m.ID,
		EnvironmentID:  m.EnvironmentID,
		TenantID:       m.TenantID,
		Channel:        m.Channel,
		TemplateID:     m.TemplateID,
		Status:         entity.JobStatus(m.Status),
		TotalCount:     m.TotalCount,
		SuccessCount:   m.SuccessCount,
		FailedCount:    m.FailedCount,
		ScheduledAt:    m.ScheduledAt,
		StartedAt:      m.StartedAt,
		CompletedAt:    m.CompletedAt,
		ErrorMessage:   m.ErrorMessage,
		Metadata:       metadata,
		RecipientsData: recipientsData,
		CreatedBy:      m.CreatedBy,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func ToJobModel(e *entity.NotificationJob) *model.NotificationJobModel {
	if e == nil {
		return nil
	}

	var metadataJSON datatypes.JSON
	if e.Metadata != nil {
		bytes, _ := json.Marshal(e.Metadata)
		metadataJSON = datatypes.JSON(bytes)
	}

	var recipientsJSON datatypes.JSON
	if e.RecipientsData != nil {
		bytes, _ := json.Marshal(e.RecipientsData)
		recipientsJSON = datatypes.JSON(bytes)
	}

	return &model.NotificationJobModel{
		ID:             e.ID,
		EnvironmentID:  e.EnvironmentID,
		TenantID:       e.TenantID,
		Channel:        e.Channel,
		TemplateID:     e.TemplateID,
		Status:         string(e.Status),
		TotalCount:     e.TotalCount,
		SuccessCount:   e.SuccessCount,
		FailedCount:    e.FailedCount,
		ScheduledAt:    e.ScheduledAt,
		StartedAt:      e.StartedAt,
		CompletedAt:    e.CompletedAt,
		ErrorMessage:   e.ErrorMessage,
		Metadata:       metadataJSON,
		RecipientsData: recipientsJSON,
		CreatedBy:      e.CreatedBy,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}

func ToJobDomainList(models []*model.NotificationJobModel) []*entity.NotificationJob {
	var result []*entity.NotificationJob
	for _, m := range models {
		result = append(result, ToJobDomain(m))
	}
	return result
}
