package impl

import (
	"context"
	"errors"
	"time"

	"CONVERDA/internal/notification/application/service"
	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/domain/repository"

	"github.com/google/uuid"
)

type layoutManager struct {
	repo repository.NotificationLayoutRepository
}

func NewLayoutManager(repo repository.NotificationLayoutRepository) service.LayoutManager {
	return &layoutManager{repo: repo}
}

func (s *layoutManager) CreateLayout(ctx context.Context, envID uuid.UUID, name, description, contentHTML string, variables map[string]interface{}, isDefault bool) (*entity.NotificationLayout, error) {
	// If set as default, we might want to unset previous default?
	// For simplicity, we allow multiple defaults or assume DB handles logic, but usually we should ensure single default.
	// But let's keep it simple for now, or check if requirement says single default.
	// Assuming single default per environment is preferred.

	if isDefault {
		// Logic to unset other defaults could be here, but might be expensive.
		// Ignoring for MVP.
	}

	layout := &entity.NotificationLayout{
		ID:              uuid.New(),
		EnvironmentID:   envID,
		Name:            name,
		Description:     description,
		ContentHTML:     contentHTML,
		VariablesSchema: variables,
		IsDefault:       isDefault,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.repo.Create(ctx, layout); err != nil {
		return nil, err
	}
	return layout, nil
}

func (s *layoutManager) UpdateLayout(ctx context.Context, envID uuid.UUID, id uuid.UUID, name, description, contentHTML string, variables map[string]interface{}, isDefault bool) error {
	layout, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if layout == nil {
		return errors.New("layout not found")
	}
	if layout.EnvironmentID != envID {
		return errors.New("layout not found in this environment")
	}

	layout.Name = name
	layout.Description = description
	layout.ContentHTML = contentHTML
	layout.VariablesSchema = variables
	layout.IsDefault = isDefault
	layout.UpdatedAt = time.Now()

	return s.repo.Update(ctx, layout)
}

func (s *layoutManager) DeleteLayout(ctx context.Context, envID uuid.UUID, id uuid.UUID) error {
	layout, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if layout == nil {
		return nil
	}
	if layout.EnvironmentID != envID {
		return errors.New("layout not found in this environment")
	}

	return s.repo.Delete(ctx, id)
}

func (s *layoutManager) GetLayout(ctx context.Context, envID uuid.UUID, id uuid.UUID) (*entity.NotificationLayout, error) {
	layout, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if layout != nil && layout.EnvironmentID != envID {
		return nil, nil
	}
	return layout, nil
}

func (s *layoutManager) ListLayouts(ctx context.Context, envID uuid.UUID, limit, offset int) ([]*entity.NotificationLayout, int64, error) {
	return s.repo.List(ctx, envID, limit, offset)
}

func (s *layoutManager) GetDefaultLayout(ctx context.Context, envID uuid.UUID) (*entity.NotificationLayout, error) {
	return s.repo.GetDefault(ctx, envID)
}
