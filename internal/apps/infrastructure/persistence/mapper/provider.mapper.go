package mapper

import (
	"encoding/json"

	"CONVERDA/internal/apps/domain/model/entity"
	"CONVERDA/internal/apps/infrastructure/persistence/model"

	"gorm.io/datatypes"
)

func ToProviderEntity(m *model.ProviderModel) *entity.Provider {
	if m == nil {
		return nil
	}
	var config map[string]interface{}
	_ = json.Unmarshal(m.Configuration, &config)

	return &entity.Provider{
		ID:            m.ID,
		TenantID:      m.TenantID,
		AppID:         m.AppID,
		EnvironmentID: m.EnvironmentID,
		ProviderType:  m.ProviderType,
		ProviderName:  m.ProviderName,
		IsActive:      m.IsActive,
		Configuration: config,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func ToProviderModel(e *entity.Provider) *model.ProviderModel {
	if e == nil {
		return nil
	}
	configJSON, _ := json.Marshal(e.Configuration)

	return &model.ProviderModel{
		ID:            e.ID,
		TenantID:      e.TenantID,
		AppID:         e.AppID,
		EnvironmentID: e.EnvironmentID,
		ProviderType:  e.ProviderType,
		ProviderName:  e.ProviderName,
		IsActive:      e.IsActive,
		Configuration: datatypes.JSON(configJSON),
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}

func ToProviderEntities(models []model.ProviderModel) []*entity.Provider {
	entities := make([]*entity.Provider, 0, len(models))
	for i := range models {
		entities = append(entities, ToProviderEntity(&models[i]))
	}
	return entities
}
