package mapper

import (
	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/infrastructure/persistence/model"
)

func ToGroupDomain(m *model.NotificationGroupModel) *entity.NotificationGroup {
	if m == nil {
		return nil
	}
	return &entity.NotificationGroup{
		ID:            m.ID,
		EnvironmentID: m.EnvironmentID,
		Name:          m.Name,
		Key:           m.Key,
		Description:   m.Description,
		IsDefault:     m.IsDefault,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func ToGroupModel(e *entity.NotificationGroup) *model.NotificationGroupModel {
	if e == nil {
		return nil
	}
	return &model.NotificationGroupModel{
		ID:            e.ID,
		EnvironmentID: e.EnvironmentID,
		Name:          e.Name,
		Key:           e.Key,
		Description:   e.Description,
		IsDefault:     e.IsDefault,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}

func ToGroupDomainList(models []*model.NotificationGroupModel) []*entity.NotificationGroup {
	var result []*entity.NotificationGroup
	for _, m := range models {
		result = append(result, ToGroupDomain(m))
	}
	return result
}
