package mapper

import (
	"encoding/json"

	"CONVERDA/internal/workflow/domain/model/entity"
	"CONVERDA/internal/workflow/infrastructure/persistence/model"

	"gorm.io/datatypes"
)

// --- Workflow ---

func ToWorkflowDomain(m *model.WorkflowModel) *entity.Workflow {
	if m == nil {
		return nil
	}
	steps := make([]entity.WorkflowStep, 0, len(m.Steps))
	for i := range m.Steps {
		steps = append(steps, *ToStepDomain(&m.Steps[i]))
	}
	return &entity.Workflow{
		ID:                m.ID,
		EnvironmentID:     m.EnvironmentID,
		Name:              m.Name,
		TriggerIdentifier: m.TriggerIdentifier,
		IsActive:          m.IsActive,
		Steps:             steps,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}

func ToWorkflowModel(e *entity.Workflow) *model.WorkflowModel {
	if e == nil {
		return nil
	}
	return &model.WorkflowModel{
		ID:                e.ID,
		EnvironmentID:     e.EnvironmentID,
		Name:              e.Name,
		TriggerIdentifier: e.TriggerIdentifier,
		IsActive:          e.IsActive,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
	}
}

// --- WorkflowStep ---

func ToStepDomain(m *model.WorkflowStepModel) *entity.WorkflowStep {
	if m == nil {
		return nil
	}
	var config map[string]interface{}
	if len(m.Config) > 0 {
		_ = json.Unmarshal(m.Config, &config)
	}
	return &entity.WorkflowStep{
		ID:           m.ID,
		WorkflowID:   m.WorkflowID,
		ParentStepID: m.ParentStepID,
		StepType:     m.StepType,
		Config:       config,
		Order:        m.Order,
		CreatedAt:    m.CreatedAt,
	}
}

func ToStepModel(e *entity.WorkflowStep) *model.WorkflowStepModel {
	if e == nil {
		return nil
	}
	configJSON, _ := json.Marshal(e.Config)
	return &model.WorkflowStepModel{
		ID:           e.ID,
		WorkflowID:   e.WorkflowID,
		ParentStepID: e.ParentStepID,
		StepType:     e.StepType,
		Config:       datatypes.JSON(configJSON),
		Order:        e.Order,
		CreatedAt:    e.CreatedAt,
	}
}
