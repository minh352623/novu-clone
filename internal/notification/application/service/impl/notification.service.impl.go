package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"CONVERDA/internal/notification/application/service"
	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/domain/repository"
	"CONVERDA/internal/notification/infrastructure/provider"

	"github.com/google/uuid"
)

type notificationServiceImpl struct {
	notifRepo   repository.NotificationRepository
	tmplManager *service.TemplateManager
	dispatcher  *provider.Dispatcher
}

func NewNotificationService(
	notifRepo repository.NotificationRepository,
	tmplManager *service.TemplateManager,
	dispatcher *provider.Dispatcher,
) service.NotificationService {
	return &notificationServiceImpl{
		notifRepo:   notifRepo,
		tmplManager: tmplManager,
		dispatcher:  dispatcher,
	}
}

func (s *notificationServiceImpl) Send(ctx context.Context, req service.SendRequest) (*service.SendResponse, error) {
	// 1. Validate inputs (basic)
	if req.TemplateCode == "" || req.Recipient == "" || req.Channel == "" {
		return nil, fmt.Errorf("missing required fields")
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}
	envID, err := uuid.Parse(req.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid environment_id: %w", err)
	}

	// 2. Compile Template
	// Default language if not provided
	lang := req.Language
	if lang == "" {
		lang = "en"
	}

	subject, body, err := s.tmplManager.GetAndCompile(ctx, envID, req.TemplateCode, lang, req.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to compile template: %w", err)
	}

	// 3. Create Notification Record (Pending)
	dataBytes, _ := json.Marshal(req.Data)
	notif := entity.NewNotification(tenantID, envID, req.TemplateCode, req.Recipient, req.Channel, dataBytes)

	if err := s.notifRepo.Create(ctx, notif); err != nil {
		return nil, fmt.Errorf("failed to create notification record: %w", err)
	}

	// 4. Dispatch
	// Convert req.Data (map[string]interface{}) to map[string]string for PushProvider
	pushData := make(map[string]string)
	for k, v := range req.Data {
		pushData[k] = fmt.Sprintf("%v", v)
	}

	dispatchErr := s.dispatcher.Dispatch(ctx, envID, req.Channel, req.Recipient, subject, body, pushData)

	// 5. Update Status
	now := time.Now()
	if dispatchErr != nil {
		notif.Status = "failed"
		errMsg := dispatchErr.Error()
		notif.ErrorMessage = &errMsg
	} else {
		notif.Status = "sent"
		notif.SentAt = &now
	}
	notif.UpdatedAt = now

	if err := s.notifRepo.Update(ctx, notif); err != nil {
		// Log error but don't fail the request if dispatch succeeded?
		// Ideally we should return error or alert.
		fmt.Printf("Failed to update notification status: %v\n", err)
	}

	if dispatchErr != nil {
		return nil, dispatchErr
	}

	return &service.SendResponse{
		NotificationID: notif.ID.String(),
		Status:         notif.Status,
	}, nil
}
