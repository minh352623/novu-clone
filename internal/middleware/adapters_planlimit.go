package middleware

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ResourceCounterAdapter implements planlimit.ResourceCounter using GORM queries
type ResourceCounterAdapter struct {
	db *gorm.DB
}

// NewResourceCounterAdapter creates a new resource counter
func NewResourceCounterAdapter(db *gorm.DB) *ResourceCounterAdapter {
	return &ResourceCounterAdapter{db: db}
}

// CountApps returns the number of apps owned by a tenant
func (a *ResourceCounterAdapter) CountApps(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var count int64
	err := a.db.WithContext(ctx).Table("apps").Where("tenant_id = ?", tenantID).Count(&count).Error
	return int(count), err
}

// CountMembers returns the number of members in a tenant
func (a *ResourceCounterAdapter) CountMembers(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var count int64
	err := a.db.WithContext(ctx).Table("tenant_members").Where("tenant_id = ?", tenantID).Count(&count).Error
	return int(count), err
}

// CountWorkflows returns the number of workflows owned by a tenant
func (a *ResourceCounterAdapter) CountWorkflows(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var count int64
	err := a.db.WithContext(ctx).Table("workflows").Where("tenant_id = ?", tenantID).Count(&count).Error
	return int(count), err
}
