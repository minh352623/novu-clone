package middleware

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantPlanRPMAdapter implements TenantPlanRPMProvider by looking up
// the tenant's pricing plan and returning its rate_limit_rpm.
type TenantPlanRPMAdapter struct {
	db *gorm.DB
}

// NewTenantPlanRPMAdapter creates a new adapter
func NewTenantPlanRPMAdapter(db *gorm.DB) *TenantPlanRPMAdapter {
	return &TenantPlanRPMAdapter{db: db}
}

// GetTenantPlanRPM looks up the RPM limit from the tenant's pricing plan.
// Returns 0 (no limit) if tenant or plan not found.
func (a *TenantPlanRPMAdapter) GetTenantPlanRPM(ctx context.Context, tenantID uuid.UUID) int {
	var rpm int
	err := a.db.WithContext(ctx).
		Table("tenants t").
		Joins("JOIN pricing_plans pp ON pp.id = t.pricing_plan_id").
		Where("t.id = ?", tenantID).
		Select("pp.rate_limit_rpm").
		Scan(&rpm).Error

	if err != nil {
		slog.Warn("TenantPlanRPMAdapter: failed to lookup RPM",
			"tenant_id", tenantID, "error", err)
		return 0
	}

	return rpm
}
