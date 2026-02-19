package test

import (
	"context"
	"testing"

	"CONVERDA/internal/notification/application/service"
	"CONVERDA/internal/notification/application/service/impl"
	"CONVERDA/internal/notification/infrastructure/provider"
	"CONVERDA/internal/notification/stub"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type noopProvider struct{}

func (p *noopProvider) Send(ctx context.Context, to, subject, body string) error {
	return nil
}

type noopPushProvider struct{}

func (p *noopPushProvider) Send(ctx context.Context, to, subject, body string, data map[string]string) error {
	return nil
}

func HelperInitService(emailProv provider.EmailProvider, pushProv provider.PushProvider) service.NotificationService {
	// 1. Repositories (Stub)
	notifRepo := stub.NewStubNotificationRepository()
	tmplRepo := stub.NewStubTemplateRepository()
	configRepo := stub.NewStubProviderConfigRepository()
	layoutRepo := stub.NewStubNotificationLayoutRepository()

	// 2. Domain Services
	tmplManager := service.NewTemplateManager(tmplRepo, layoutRepo)
	dispatcher := provider.NewDispatcher(configRepo)
	if emailProv != nil {
		dispatcher.SetEmailProvider(emailProv)
	}
	if pushProv != nil {
		dispatcher.SetPushProvider(pushProv)
	}

	// 3. Application Service
	notifUoW := stub.NewStubNotificationUnitOfWork(notifRepo)
	return impl.NewNotificationService(notifRepo, notifUoW, tmplManager, dispatcher)
}

func TestNotificationFlow(t *testing.T) {
	// Initialize Module with Stubs & Noop Providers
	svc := HelperInitService(&noopProvider{}, &noopPushProvider{})

	// Test Case 1: Send Email Success
	req := service.SendRequest{
		TenantID:      uuid.New().String(),
		EnvironmentID: uuid.New().String(),
		TemplateCode:  "WELCOME_EMAIL",
		Recipient:     "user@example.com",
		Channel:       "smtp",
		Data: map[string]interface{}{
			"Name": "John Doe",
		},
		Language: "en",
	}

	resp, err := svc.Send(context.Background(), req)
	if !assert.NoError(t, err) {
		return
	}
	assert.NotNil(t, resp)
	assert.Equal(t, "sent", resp.Status)
	assert.NotEmpty(t, resp.NotificationID)

	// Test Case 2: Send Push (FCM) Success with Data
	reqPush := service.SendRequest{
		TenantID:      uuid.New().String(),
		EnvironmentID: uuid.New().String(),
		TemplateCode:  "NEW_MESSAGE",
		Recipient:     "device_token_123",
		Channel:       "fcm",
		Data: map[string]interface{}{
			"ThreadID": "thread_abc",
			"Message":  "Hello world",
		},
		Language: "en",
	}

	respPush, err := svc.Send(context.Background(), reqPush)
	assert.NoError(t, err)
	assert.NotNil(t, respPush)
	assert.Equal(t, "sent", respPush.Status)

	// Test Case 2: Missing Fields
	reqInvalid := service.SendRequest{
		Channel: "smtp",
	}
	_, err = svc.Send(context.Background(), reqInvalid)
	assert.Error(t, err)
}
