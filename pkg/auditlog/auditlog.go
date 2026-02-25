// Package auditlog provides structured audit logging for multi-tenant compliance.
// It records who did what, to which resource, and when — scoped to a tenant.
package auditlog

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"CONVERDA/pkg/tenantctx"

	"github.com/google/uuid"
)

// Entry represents a single audit log record
type Entry struct {
	ID           uuid.UUID       `json:"id"`
	TenantID     uuid.UUID       `json:"tenant_id"`
	ActorID      *uuid.UUID      `json:"actor_id,omitempty"`
	ActorType    string          `json:"actor_type"` // user, system, api_key
	Action       string          `json:"action"`     // create, update, delete, login, export
	ResourceType string          `json:"resource_type"`
	ResourceID   string          `json:"resource_id,omitempty"`
	Changes      json.RawMessage `json:"changes,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

// Store persists audit log entries
type Store interface {
	Save(ctx context.Context, entry *Entry) error
	List(ctx context.Context, filter Filter) ([]*Entry, int64, error)
}

// Filter for querying audit logs
type Filter struct {
	TenantID     uuid.UUID
	ActorID      *uuid.UUID
	Action       string
	ResourceType string
	ResourceID   string
	From         *time.Time
	To           *time.Time
	Limit        int
	Offset       int
}

// Logger wraps Store and provides a convenient API for recording audit events.
type Logger struct {
	store Store
}

// NewLogger creates a new audit logger
func NewLogger(store Store) *Logger {
	return &Logger{store: store}
}

// Log records an audit event. It extracts tenant context automatically.
func (l *Logger) Log(ctx context.Context, action, resourceType, resourceID string, opts ...Option) {
	tenantID := tenantctx.TenantIDFromCtx(ctx)
	if tenantID == uuid.Nil {
		slog.Warn("auditlog: skipping — no tenant_id in context", "action", action)
		return
	}

	entry := &Entry{
		ID:           uuid.New(),
		TenantID:     tenantID,
		ActorType:    "system",
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		CreatedAt:    time.Now(),
	}

	for _, opt := range opts {
		opt(entry)
	}

	if err := l.store.Save(ctx, entry); err != nil {
		slog.Error("auditlog: failed to save", "error", err, "action", action, "resource", resourceType)
	}
}

// Option configures an audit log entry
type Option func(*Entry)

// WithActor sets the actor (user who performed the action)
func WithActor(actorID uuid.UUID, actorType string) Option {
	return func(e *Entry) {
		e.ActorID = &actorID
		e.ActorType = actorType
	}
}

// WithChanges records before/after changes
func WithChanges(changes map[string]interface{}) Option {
	return func(e *Entry) {
		data, _ := json.Marshal(changes)
		e.Changes = data
	}
}

// WithMetadata adds extra metadata (IP, user-agent, etc.)
func WithMetadata(meta map[string]interface{}) Option {
	return func(e *Entry) {
		data, _ := json.Marshal(meta)
		e.Metadata = data
	}
}

// Standard actions
const (
	ActionCreate     = "create"
	ActionUpdate     = "update"
	ActionDelete     = "delete"
	ActionLogin      = "login"
	ActionLogout     = "logout"
	ActionExport     = "export"
	ActionDataErase  = "data_erase"  // GDPR
	ActionDataExport = "data_export" // GDPR
)
