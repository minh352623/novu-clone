package mapper

import (
	"encoding/json"

	"CONVERDA/internal/workflow/domain/model/entity"
	"CONVERDA/internal/workflow/infrastructure/persistence/model"

	"gorm.io/datatypes"
)

// --- WorkflowExecution ---

func ToExecutionDomain(m *model.WorkflowExecutionModel) *entity.WorkflowExecution {
	if m == nil {
		return nil
	}
	var payload map[string]interface{}
	if len(m.TriggerPayload) > 0 {
		_ = json.Unmarshal(m.TriggerPayload, &payload)
	}
	return &entity.WorkflowExecution{
		ID:             m.ID,
		WorkflowID:     m.WorkflowID,
		SubscriberKey:  m.SubscriberKey,
		TriggerPayload: payload,
		Status:         m.Status,
		CurrentStepID:  m.CurrentStepID,
		StartedAt:      m.StartedAt,
		CompletedAt:    m.CompletedAt,
	}
}

func ToExecutionModel(e *entity.WorkflowExecution) *model.WorkflowExecutionModel {
	if e == nil {
		return nil
	}
	payloadJSON, _ := json.Marshal(e.TriggerPayload)
	return &model.WorkflowExecutionModel{
		ID:             e.ID,
		WorkflowID:     e.WorkflowID,
		SubscriberKey:  e.SubscriberKey,
		TriggerPayload: datatypes.JSON(payloadJSON),
		Status:         e.Status,
		CurrentStepID:  e.CurrentStepID,
		StartedAt:      e.StartedAt,
		CompletedAt:    e.CompletedAt,
	}
}

// --- StepExecution ---

func ToStepExecutionDomain(m *model.StepExecutionModel) *entity.StepExecution {
	if m == nil {
		return nil
	}
	var output map[string]interface{}
	if len(m.Output) > 0 {
		_ = json.Unmarshal(m.Output, &output)
	}
	return &entity.StepExecution{
		ID:          m.ID,
		ExecutionID: m.ExecutionID,
		StepID:      m.StepID,
		Status:      m.Status,
		ScheduledAt: m.ScheduledAt,
		StartedAt:   m.StartedAt,
		CompletedAt: m.CompletedAt,
		Output:      output,
	}
}

func ToStepExecutionModel(e *entity.StepExecution) *model.StepExecutionModel {
	if e == nil {
		return nil
	}
	outputJSON, _ := json.Marshal(e.Output)
	return &model.StepExecutionModel{
		ID:          e.ID,
		ExecutionID: e.ExecutionID,
		StepID:      e.StepID,
		Status:      e.Status,
		ScheduledAt: e.ScheduledAt,
		StartedAt:   e.StartedAt,
		CompletedAt: e.CompletedAt,
		Output:      datatypes.JSON(outputJSON),
	}
}

// --- DigestEvent ---

func ToDigestEventDomain(m *model.DigestEventModel) *entity.DigestEvent {
	if m == nil {
		return nil
	}
	var payload map[string]interface{}
	if len(m.Payload) > 0 {
		_ = json.Unmarshal(m.Payload, &payload)
	}
	return &entity.DigestEvent{
		ID:            m.ID,
		StepID:        m.StepID,
		ExecutionID:   m.ExecutionID,
		SubscriberKey: m.SubscriberKey,
		Payload:       payload,
		CreatedAt:     m.CreatedAt,
	}
}

func ToDigestEventModel(e *entity.DigestEvent) *model.DigestEventModel {
	if e == nil {
		return nil
	}
	payloadJSON, _ := json.Marshal(e.Payload)
	return &model.DigestEventModel{
		ID:            e.ID,
		StepID:        e.StepID,
		ExecutionID:   e.ExecutionID,
		SubscriberKey: e.SubscriberKey,
		Payload:       datatypes.JSON(payloadJSON),
		CreatedAt:     e.CreatedAt,
	}
}
