package mapper

import (
	"CONVERDA/internal/apps/domain/model/entity"
	"CONVERDA/internal/apps/infrastructure/persistence/model"
)

func ToSystemEnvironmentEntity(m *model.SystemEnvironmentModel) *entity.SystemEnvironment {
	if m == nil {
		return nil
	}
	return &entity.SystemEnvironment{
		Code:        m.Code,
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
	}
}

func ToSystemEnvironmentEntities(models []model.SystemEnvironmentModel) []*entity.SystemEnvironment {
	entities := make([]*entity.SystemEnvironment, len(models))
	for i, m := range models {
		entities[i] = ToSystemEnvironmentEntity(&m)
	}
	return entities
}
