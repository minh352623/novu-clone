package mapper

import (
	"encoding/json"

	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/infrastructure/persistence/model"

	"gorm.io/datatypes"
)

func ToLayoutDomain(m *model.NotificationLayoutModel) *entity.NotificationLayout {
	if m == nil {
		return nil
	}

	var schema map[string]interface{}
	if len(m.VariablesSchema) > 0 {
		_ = json.Unmarshal(m.VariablesSchema, &schema)
	}

	return &entity.NotificationLayout{
		ID:              m.ID,
		EnvironmentID:   m.EnvironmentID,
		Name:            m.Name,
		Description:     m.Description,
		ContentHTML:     m.ContentHTML,
		VariablesSchema: schema,
		IsDefault:       m.IsDefault,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func ToLayoutModel(e *entity.NotificationLayout) *model.NotificationLayoutModel {
	if e == nil {
		return nil
	}

	var schemaJSON datatypes.JSON
	if e.VariablesSchema != nil {
		bytes, _ := json.Marshal(e.VariablesSchema)
		schemaJSON = datatypes.JSON(bytes)
	} else {
		schemaJSON = datatypes.JSON("{}")
	}

	return &model.NotificationLayoutModel{
		ID:              e.ID,
		EnvironmentID:   e.EnvironmentID,
		Name:            e.Name,
		Description:     e.Description,
		ContentHTML:     e.ContentHTML,
		VariablesSchema: schemaJSON,
		IsDefault:       e.IsDefault,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}

func ToLayoutDomainList(models []*model.NotificationLayoutModel) []*entity.NotificationLayout {
	var result []*entity.NotificationLayout
	for _, m := range models {
		result = append(result, ToLayoutDomain(m))
	}
	return result
}
