package mapper

import (
	"CONVERDA/internal/apps/domain/model/entity"
	"CONVERDA/internal/apps/infrastructure/persistence/model"
)

func ToAppEntity(m *model.AppModel) *entity.App {
	if m == nil {
		return nil
	}
	app := &entity.App{
		ID:          m.ID,
		TenantID:    m.TenantID,
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}

	if len(m.Environments) > 0 {
		app.Environments = make([]*entity.Environment, len(m.Environments))
		for i, env := range m.Environments {
			app.Environments[i] = ToEnvironmentEntity(&env)
		}
	}

	return app
}

func ToAppModel(e *entity.App) *model.AppModel {
	if e == nil {
		return nil
	}
	return &model.AppModel{
		ID:          e.ID,
		TenantID:    e.TenantID,
		Name:        e.Name,
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func ToEnvironmentEntity(m *model.EnvironmentModel) *entity.Environment {
	if m == nil {
		return nil
	}
	return &entity.Environment{
		ID:              m.ID,
		AppID:           m.AppID,
		EnvironmentCode: m.EnvironmentCode,
		APIKey:          m.APIKey,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func ToEnvironmentModel(e *entity.Environment) *model.EnvironmentModel {
	if e == nil {
		return nil
	}
	return &model.EnvironmentModel{
		ID:              e.ID,
		AppID:           e.AppID,
		EnvironmentCode: e.EnvironmentCode,
		APIKey:          e.APIKey,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}

func ToAppEntities(models []model.AppModel) []*entity.App {
	entities := make([]*entity.App, len(models))
	for i, m := range models {
		entities[i] = ToAppEntity(&m)
	}
	return entities
}

func ToEnvironmentEntities(models []model.EnvironmentModel) []*entity.Environment {
	entities := make([]*entity.Environment, len(models))
	for i, m := range models {
		entities[i] = ToEnvironmentEntity(&m)
	}
	return entities
}
