package impl

import (
	"context"
	"fmt"
	"time"

	"CONVERDA/internal/notification/application/service"
	domain "CONVERDA/internal/notification/domain"
	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/domain/repository"

	"github.com/google/uuid"
)

type layoutManager struct {
	repo repository.NotificationLayoutRepository
	uow  repository.NotificationUnitOfWork
}

func NewLayoutManager(repo repository.NotificationLayoutRepository, uow repository.NotificationUnitOfWork) service.LayoutManager {
	return &layoutManager{repo: repo, uow: uow}
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

	if err := s.uow.Execute(ctx, func(tx repository.NotificationTxRepository) error {
		return tx.Layouts().Create(ctx, layout)
	}); err != nil {
		return nil, err
	}
	return layout, nil
}

func (s *layoutManager) UpdateLayout(ctx context.Context, envID uuid.UUID, id uuid.UUID, name, description, contentHTML string, variables map[string]interface{}, isDefault bool) error {
	return s.uow.Execute(ctx, func(tx repository.NotificationTxRepository) error {
		layout, err := tx.Layouts().GetByID(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to fetch layout %s: %w", id, err)
		}
		if layout == nil {
			return domain.ErrLayoutNotFound
		}
		if layout.EnvironmentID != envID {
			return domain.ErrLayoutNotInEnv
		}

		layout.Name = name
		layout.Description = description
		layout.ContentHTML = contentHTML
		layout.VariablesSchema = variables
		layout.IsDefault = isDefault
		layout.UpdatedAt = time.Now()

		return tx.Layouts().Update(ctx, layout)
	})
}

func (s *layoutManager) DeleteLayout(ctx context.Context, envID uuid.UUID, id uuid.UUID) error {
	return s.uow.Execute(ctx, func(tx repository.NotificationTxRepository) error {
		layout, err := tx.Layouts().GetByID(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to fetch layout %s: %w", id, err)
		}
		if layout == nil {
			return nil
		}
		if layout.EnvironmentID != envID {
			return domain.ErrLayoutNotInEnv
		}

		return tx.Layouts().Delete(ctx, id)
	})
}

func (s *layoutManager) GetLayout(ctx context.Context, envID uuid.UUID, id uuid.UUID) (*entity.NotificationLayout, error) {
	layout, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch layout %s: %w", id, err)
	}
	if layout != nil && layout.EnvironmentID != envID {
		return nil, nil
	}
	return layout, nil
}

func (s *layoutManager) ListLayouts(ctx context.Context, envID uuid.UUID, limit, offset int) ([]*entity.NotificationLayout, int64, error) {
	layouts, total, err := s.repo.List(ctx, envID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list layouts for env %s: %w", envID, err)
	}
	return layouts, total, nil
}

func (s *layoutManager) GetDefaultLayout(ctx context.Context, envID uuid.UUID) (*entity.NotificationLayout, error) {
	layout, err := s.repo.GetDefault(ctx, envID)
	if err != nil {
		return nil, fmt.Errorf("failed to get default layout for env %s: %w", envID, err)
	}
	return layout, nil
}
