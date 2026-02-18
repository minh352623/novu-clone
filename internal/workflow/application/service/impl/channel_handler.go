package impl

import (
	"context"
	"fmt"

	"CONVERDA/internal/workflow/domain/model/entity"
	"CONVERDA/internal/workflow/domain/repository"

	"github.com/google/uuid"
)

// ChannelHandler sends a notification via the NotificationService.
type ChannelHandler struct {
	notifier repository.Notifier
}

func NewChannelHandler(notifier repository.Notifier) *ChannelHandler {
	return &ChannelHandler{notifier: notifier}
}

func (h *ChannelHandler) Execute(ctx context.Context, step *entity.WorkflowStep,
	exec *entity.WorkflowExecution, stepExec *entity.StepExecution) error {

	// Extract config: { "template_code": "welcome", "channel": "email" }
	templateCode, _ := step.Config["template_code"].(string)
	channel, _ := step.Config["channel"].(string)

	if templateCode == "" {
		return fmt.Errorf("channel step %s: missing template_code in config", step.ID)
	}
	if channel == "" {
		channel = "email" // default
	}

	req := repository.NotificationRequest{
		EnvironmentID: exec.WorkflowID, // Fallback to WorkflowID if EnvironmentID not on execution (Check domain logic)
		TemplateCode:  templateCode,
		Recipient:     exec.SubscriberKey,
		Channel:       channel,
		Data:          exec.TriggerPayload,
	}

	// Environment ID override if explicitly in step config
	if envIDStr, ok := step.Config["environment_id"].(string); ok {
		if id, err := uuid.Parse(envIDStr); err == nil {
			req.EnvironmentID = id
		}
	}

	// Recipient override
	if recipient, ok := step.Config["recipient"].(string); ok {
		req.Recipient = recipient
	}

	// Language
	if lang, ok := step.Config["language"].(string); ok {
		req.Language = lang
	}

	return h.notifier.Send(ctx, req)
}
