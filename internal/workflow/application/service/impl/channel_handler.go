package impl

import (
	"context"
	"fmt"

	notifService "CONVERDA/internal/notification/application/service"
	"CONVERDA/internal/workflow/domain/model/entity"
)

// ChannelHandler sends a notification via the NotificationService.
type ChannelHandler struct {
	notifSvc notifService.NotificationService
}

func NewChannelHandler(notifSvc notifService.NotificationService) *ChannelHandler {
	return &ChannelHandler{notifSvc: notifSvc}
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

	req := notifService.SendRequest{
		EnvironmentID: exec.WorkflowID.String(), // Will be resolved from workflow
		TemplateCode:  templateCode,
		Recipient:     exec.SubscriberKey,
		Channel:       channel,
		Data:          exec.TriggerPayload,
	}

	// Override with step-level config if provided
	if tenantID, ok := step.Config["tenant_id"].(string); ok {
		req.TenantID = tenantID
	}
	if envID, ok := step.Config["environment_id"].(string); ok {
		req.EnvironmentID = envID
	}
	if lang, ok := step.Config["language"].(string); ok {
		req.Language = lang
	}

	_, err := h.notifSvc.Send(ctx, req)
	return err
}
