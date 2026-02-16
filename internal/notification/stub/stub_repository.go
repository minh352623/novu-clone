package stub

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/domain/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// InMemory Notification Repo
type StubNotificationRepository struct {
	data map[uuid.UUID]*entity.Notification
	mu   sync.Mutex
}

func NewStubNotificationRepository() *StubNotificationRepository {
	return &StubNotificationRepository{
		data: make(map[uuid.UUID]*entity.Notification),
	}
}

func (r *StubNotificationRepository) Create(ctx context.Context, notif *entity.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[notif.ID] = notif
	return nil
}

func (r *StubNotificationRepository) Update(ctx context.Context, notif *entity.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[notif.ID]; !ok {
		return errors.New("notification not found")
	}
	r.data[notif.ID] = notif
	return nil
}

func (r *StubNotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Notification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if n, ok := r.data[id]; ok {
		return n, nil
	}
	return nil, errors.New("notification not found")
}

// InMemory Template Repo
type StubTemplateRepository struct {
	mock.Mock
}

func NewStubTemplateRepository() *StubTemplateRepository {
	return &StubTemplateRepository{}
}

func (r *StubTemplateRepository) GetByCode(ctx context.Context, envID uuid.UUID, code, lang string) (*entity.Template, error) {
	return &entity.Template{
		ID:        uuid.New(),
		Code:      code,
		Version:   1,
		Subject:   "Subject: {{.Name}}",
		Body:      "Hello {{.Name}}, this is a notification.",
		Language:  lang,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (r *StubTemplateRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Template, error) {
	return &entity.Template{
		ID:        id,
		Code:      "DUMMY_CODE",
		Version:   1,
		Subject:   "Subject",
		Body:      "Body",
		Language:  "en",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// InMemory Provider Config Repo
type StubProviderConfigRepository struct{}

func NewStubProviderConfigRepository() *StubProviderConfigRepository {
	return &StubProviderConfigRepository{}
}

func (r *StubProviderConfigRepository) GetActive(ctx context.Context, envID uuid.UUID, providerType string) (*entity.ProviderConfig, error) {
	cfg := map[string]string{"host": "smtp.example.com"}
	cfgBytes, _ := json.Marshal(cfg)
	return &entity.ProviderConfig{
		ID:            uuid.New(),
		EnvironmentID: envID,
		Type:          providerType,
		Config:        cfgBytes,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
}

type stubNotificationLayoutRepository struct {
	mock.Mock
}

func NewStubNotificationLayoutRepository() repository.NotificationLayoutRepository {
	return &stubNotificationLayoutRepository{}
}

func (m *stubNotificationLayoutRepository) Create(ctx context.Context, layout *entity.NotificationLayout) error {
	return m.Called(ctx, layout).Error(0)
}

func (m *stubNotificationLayoutRepository) Update(ctx context.Context, layout *entity.NotificationLayout) error {
	return m.Called(ctx, layout).Error(0)
}

func (m *stubNotificationLayoutRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.NotificationLayout, error) {
	if id == uuid.Nil {
		return nil, nil
	}
	return &entity.NotificationLayout{
		ID:          id,
		ContentHTML: "<html><body>{{.Content}}</body></html>",
	}, nil
}

func (m *stubNotificationLayoutRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *stubNotificationLayoutRepository) List(ctx context.Context, envID uuid.UUID, limit, offset int) ([]*entity.NotificationLayout, int64, error) {
	args := m.Called(ctx, envID, limit, offset)
	return args.Get(0).([]*entity.NotificationLayout), int64(m.Called(ctx, envID, limit, offset).Int(1)), args.Error(2)
}

func (m *stubNotificationLayoutRepository) GetDefault(ctx context.Context, envID uuid.UUID) (*entity.NotificationLayout, error) {
	return &entity.NotificationLayout{
		ID:          uuid.New(),
		ContentHTML: "<html><body>{{.Content}} (Default)</body></html>",
	}, nil
}
