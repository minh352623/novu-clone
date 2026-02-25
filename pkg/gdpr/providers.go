package gdpr

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProfileProvider exports/erases user profile data (users + tenant_members)
type ProfileProvider struct {
	db *gorm.DB
}

func NewProfileProvider(db *gorm.DB) *ProfileProvider {
	return &ProfileProvider{db: db}
}

func (p *ProfileProvider) Category() DataCategory { return CategoryProfile }

func (p *ProfileProvider) Export(ctx context.Context, tenantID, userID uuid.UUID) (json.RawMessage, error) {
	var results []map[string]interface{}

	// Export user profile
	err := p.db.WithContext(ctx).
		Table("users").
		Where("id = ?", userID).
		Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("export profile: %w", err)
	}

	// Export tenant memberships
	var memberships []map[string]interface{}
	err = p.db.WithContext(ctx).
		Table("tenant_members").
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Find(&memberships).Error
	if err != nil {
		return nil, fmt.Errorf("export memberships: %w", err)
	}

	data := map[string]interface{}{
		"profile":     results,
		"memberships": memberships,
	}
	return json.Marshal(data)
}

func (p *ProfileProvider) Erase(ctx context.Context, tenantID, userID uuid.UUID) (int64, error) {
	// Anonymize user data instead of hard delete (preserve referential integrity)
	result := p.db.WithContext(ctx).
		Table("users").
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"email":     fmt.Sprintf("erased_%s@deleted.local", userID.String()[:8]),
			"full_name": nil,
		})
	return result.RowsAffected, result.Error
}

// MessageProvider exports/erases user messages
type MessageProvider struct {
	db *gorm.DB
}

func NewMessageProvider(db *gorm.DB) *MessageProvider {
	return &MessageProvider{db: db}
}

func (p *MessageProvider) Category() DataCategory { return CategoryMessages }

func (p *MessageProvider) Export(ctx context.Context, tenantID, userID uuid.UUID) (json.RawMessage, error) {
	var results []map[string]interface{}
	err := p.db.WithContext(ctx).
		Table("messages m").
		Joins("JOIN threads t ON t.id = m.thread_id").
		Where("t.tenant_id = ? AND m.sender_type = ? AND m.sender_id = ?", tenantID, "user", userID).
		Select("m.*").
		Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("export messages: %w", err)
	}
	return json.Marshal(results)
}

func (p *MessageProvider) Erase(ctx context.Context, tenantID, userID uuid.UUID) (int64, error) {
	// Redact message content instead of hard delete
	result := p.db.WithContext(ctx).
		Table("messages").
		Where("sender_type = ? AND sender_id = ? AND thread_id IN (?)",
			"user", userID,
			p.db.Table("threads").Select("id").Where("tenant_id = ?", tenantID),
		).
		Update("content", "[REDACTED per GDPR request]")
	return result.RowsAffected, result.Error
}

// SubscriptionProvider exports/erases subscriber data
type SubscriptionProvider struct {
	db *gorm.DB
}

func NewSubscriptionProvider(db *gorm.DB) *SubscriptionProvider {
	return &SubscriptionProvider{db: db}
}

func (p *SubscriptionProvider) Category() DataCategory { return CategorySubscriptions }

func (p *SubscriptionProvider) Export(ctx context.Context, tenantID, userID uuid.UUID) (json.RawMessage, error) {
	var results []map[string]interface{}
	err := p.db.WithContext(ctx).
		Table("subscribers").
		Where("tenant_id = ? AND subscriber_key = ?", tenantID, userID.String()).
		Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("export subscribers: %w", err)
	}
	return json.Marshal(results)
}

func (p *SubscriptionProvider) Erase(ctx context.Context, tenantID, userID uuid.UUID) (int64, error) {
	// Anonymize subscriber data
	result := p.db.WithContext(ctx).
		Table("subscribers").
		Where("tenant_id = ? AND subscriber_key = ?", tenantID, userID.String()).
		Updates(map[string]interface{}{
			"email": nil,
			"phone": nil,
			"data":  "{}",
		})
	return result.RowsAffected, result.Error
}

// AuditLogProvider exports audit log data (read-only, no erase — compliance requirement)
type AuditLogProvider struct {
	db *gorm.DB
}

func NewAuditLogProvider(db *gorm.DB) *AuditLogProvider {
	return &AuditLogProvider{db: db}
}

func (p *AuditLogProvider) Category() DataCategory { return CategoryAuditLogs }

func (p *AuditLogProvider) Export(ctx context.Context, tenantID, userID uuid.UUID) (json.RawMessage, error) {
	var results []map[string]interface{}
	err := p.db.WithContext(ctx).
		Table("audit_logs").
		Where("tenant_id = ? AND actor_id = ?", tenantID, userID).
		Order("created_at DESC").
		Limit(1000). // Cap for safety
		Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("export audit logs: %w", err)
	}
	return json.Marshal(results)
}

func (p *AuditLogProvider) Erase(_ context.Context, _, _ uuid.UUID) (int64, error) {
	// Audit logs must NOT be erased — compliance/legal requirement
	return 0, nil
}
