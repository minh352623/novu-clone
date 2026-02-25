package repository

import (
	"context"

	"CONVERDA/pkg/auditlog"

	"gorm.io/gorm"
)

// AuditLogModel is the GORM model for audit_logs table
type AuditLogModel struct {
	ID           string  `gorm:"column:id;type:uuid;primaryKey"`
	TenantID     string  `gorm:"column:tenant_id;type:uuid;not null"`
	ActorID      *string `gorm:"column:actor_id;type:uuid"`
	ActorType    string  `gorm:"column:actor_type;type:text;not null"`
	Action       string  `gorm:"column:action;type:text;not null"`
	ResourceType string  `gorm:"column:resource_type;type:text;not null"`
	ResourceID   string  `gorm:"column:resource_id;type:text"`
	Changes      []byte  `gorm:"column:changes;type:jsonb;default:'{}'"`
	Metadata     []byte  `gorm:"column:metadata;type:jsonb;default:'{}'"`
	CreatedAt    string  `gorm:"column:created_at;autoCreateTime"`
}

func (AuditLogModel) TableName() string {
	return "audit_logs"
}

// AuditLogRepository implements auditlog.Store using GORM
type AuditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Save persists an audit log entry
func (r *AuditLogRepository) Save(ctx context.Context, entry *auditlog.Entry) error {
	var actorID *string
	if entry.ActorID != nil {
		s := entry.ActorID.String()
		actorID = &s
	}

	model := AuditLogModel{
		ID:           entry.ID.String(),
		TenantID:     entry.TenantID.String(),
		ActorID:      actorID,
		ActorType:    entry.ActorType,
		Action:       entry.Action,
		ResourceType: entry.ResourceType,
		ResourceID:   entry.ResourceID,
		Changes:      entry.Changes,
		Metadata:     entry.Metadata,
	}

	return r.db.WithContext(ctx).Create(&model).Error
}

// List queries audit logs with filtering
func (r *AuditLogRepository) List(ctx context.Context, filter auditlog.Filter) ([]*auditlog.Entry, int64, error) {
	// Placeholder — full implementation requires parsing back to Entry
	// For now, return empty to satisfy interface
	return nil, 0, nil
}
