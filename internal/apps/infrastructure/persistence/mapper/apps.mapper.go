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
		ID:                  m.ID,
		TenantID:            m.TenantID,
		Name:                m.Name,
		Description:         m.Description,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
		SLAThresholdSeconds: m.SLAThresholdSeconds,
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
		ID:                  e.ID,
		TenantID:            e.TenantID,
		Name:                e.Name,
		Description:         e.Description,
		CreatedAt:           e.CreatedAt,
		UpdatedAt:           e.UpdatedAt,
		SLAThresholdSeconds: e.SLAThresholdSeconds,
	}
}

func ToEnvironmentEntity(m *model.EnvironmentModel) *entity.Environment {
	if m == nil {
		return nil
	}
	env := &entity.Environment{
		ID:                  m.ID,
		AppID:               m.AppID,
		EnvironmentCode:     m.EnvironmentCode,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
		SLAThresholdSeconds: m.SLAThresholdSeconds,
		RateLimitRPM:        m.RateLimitRPM,
		RateLimitDaily:      m.RateLimitDaily,
	}

	if len(m.Keys) > 0 {
		env.Keys = make([]*entity.APIKey, len(m.Keys))
		for i, k := range m.Keys {
			env.Keys[i] = ToAPIKeyEntity(&k)
		}
	}

	return env
}

func ToEnvironmentModel(e *entity.Environment) *model.EnvironmentModel {
	if e == nil {
		return nil
	}
	return &model.EnvironmentModel{
		ID:                  e.ID,
		AppID:               e.AppID,
		EnvironmentCode:     e.EnvironmentCode,
		CreatedAt:           e.CreatedAt,
		UpdatedAt:           e.UpdatedAt,
		SLAThresholdSeconds: e.SLAThresholdSeconds,
		RateLimitRPM:        e.RateLimitRPM,
		RateLimitDaily:      e.RateLimitDaily,
	}
}

func ToAPIKeyEntity(m *model.APIKeyModel) *entity.APIKey {
	if m == nil {
		return nil
	}
	return &entity.APIKey{
		ID:            m.ID,
		AppID:         m.AppID,
		EnvironmentID: m.EnvironmentID,
		KeyHash:       m.KeyHash,
		KeyPrefix:     m.KeyPrefix,
		KeySuffix:     m.KeySuffix,
		Name:          m.Name,
		ExpiresAt:     m.ExpiresAt,
		RevokedAt:     m.RevokedAt,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func ToAPIKeyModel(e *entity.APIKey) *model.APIKeyModel {
	if e == nil {
		return nil
	}
	return &model.APIKeyModel{
		ID:            e.ID,
		AppID:         e.AppID,
		EnvironmentID: e.EnvironmentID,
		KeyHash:       e.KeyHash,
		KeyPrefix:     e.KeyPrefix,
		KeySuffix:     e.KeySuffix,
		Name:          e.Name,
		ExpiresAt:     e.ExpiresAt,
		RevokedAt:     e.RevokedAt,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
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
