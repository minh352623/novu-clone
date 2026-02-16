package mapper

import (
	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/infrastructure/persistence/model"
)

func ToProviderConfigDomain(pc *model.ProviderConfigModel) *entity.ProviderConfig {
	return &entity.ProviderConfig{
		ID:            pc.ID,
		EnvironmentID: pc.EnvironmentID,
		Type:          pc.Provider.ProviderType, // Joined
		Config:        pc.Configuration,
		IsActive:      pc.IsActive,
		CreatedAt:     pc.CreatedAt,
		UpdatedAt:     pc.UpdatedAt,
	}
}
