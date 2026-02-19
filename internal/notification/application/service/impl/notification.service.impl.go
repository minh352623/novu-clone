package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"CONVERDA/global"
	"CONVERDA/internal/notification/application/service"
	notifDomain "CONVERDA/internal/notification/domain"
	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/domain/repository"

	"github.com/google/uuid"
)

type notificationServiceImpl struct {
	notifRepo   repository.NotificationRepository
	uow         repository.NotificationUnitOfWork
	tmplManager *service.TemplateManager
	dispatcher  repository.INotificationDispatcher
}

func NewNotificationService(
	notifRepo repository.NotificationRepository,
	uow repository.NotificationUnitOfWork,
	tmplManager *service.TemplateManager,
	dispatcher repository.INotificationDispatcher,
) service.NotificationService {
	return &notificationServiceImpl{
		notifRepo:   notifRepo,
		uow:         uow,
		tmplManager: tmplManager,
		dispatcher:  dispatcher,
	}
}

func (s *notificationServiceImpl) Send(ctx context.Context, req service.SendRequest) (*service.SendResponse, error) {
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
		lang = notifDomain.DefaultLanguage
	}

	subject, body, err := s.tmplManager.GetAndCompile(ctx, envID, req.TemplateCode, lang, req.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to compile template: %w", err)
	}

	// 3. Create Notification Record (Pending)
	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		global.Logger.Warn("notification_service: failed to marshal notification data", "error", err)
	}
	notif := entity.NewNotification(tenantID, envID, req.TemplateCode, req.Recipient, req.Channel, dataBytes)

	if err := s.uow.Execute(ctx, func(tx repository.NotificationTxRepository) error {
		return tx.Notifications().Create(ctx, notif)
	}); err != nil {
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
	if err := s.uow.Execute(ctx, func(tx repository.NotificationTxRepository) error {
		now := time.Now()
		if dispatchErr != nil {
			notif.Status = notifDomain.StatusFailed
			errMsg := dispatchErr.Error()
			notif.ErrorMessage = &errMsg
		} else {
			notif.Status = notifDomain.StatusSent
			notif.SentAt = &now
		}
		notif.UpdatedAt = now

		return tx.Notifications().Update(ctx, notif)
	}); err != nil {
		global.Logger.Error("notification_service: failed to update notification status", "notificationID", notif.ID.String(), "error", err)
	}

	if dispatchErr != nil {
		return nil, dispatchErr
	}

	return &service.SendResponse{
		NotificationID: notif.ID.String(),
		Status:         notif.Status,
	}, nil
}
