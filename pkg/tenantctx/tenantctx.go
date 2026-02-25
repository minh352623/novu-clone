// Package tenantctx provides helpers for tenant context propagation.
// All services and repositories should use these helpers to extract
// tenant information from context, ensuring consistent multi-tenant isolation.
package tenantctx

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

type contextKey string

const (
	// TenantIDKey is the standard context key for tenant ID
	TenantIDKey contextKey = "tenant_id"
	// TenantSlugKey is the standard context key for tenant slug
	TenantSlugKey contextKey = "tenant_slug"
	// TenantPlanKey is the standard context key for tenant plan
	TenantPlanKey contextKey = "tenant_plan"
)

// WithTenantID returns a new context with the given tenant ID
func WithTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// WithTenantSlug returns a new context with the given tenant slug
func WithTenantSlug(ctx context.Context, slug string) context.Context {
	return context.WithValue(ctx, TenantSlugKey, slug)
}

// WithTenantPlan returns a new context with the given plan name
func WithTenantPlan(ctx context.Context, plan string) context.Context {
	return context.WithValue(ctx, TenantPlanKey, plan)
}

// TenantIDFromCtx extracts tenant ID from context, returns uuid.Nil if not found
func TenantIDFromCtx(ctx context.Context) uuid.UUID {
	val := ctx.Value(TenantIDKey)
	if val == nil {
		return uuid.Nil
	}
	switch v := val.(type) {
	case uuid.UUID:
		return v
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			return uuid.Nil
		}
		return id
	default:
		return uuid.Nil
	}
}

// MustTenantIDFromCtx extracts tenant ID from context, panics if not found.
// Use only in code paths where tenant context is guaranteed (after middleware).
func MustTenantIDFromCtx(ctx context.Context) uuid.UUID {
	id := TenantIDFromCtx(ctx)
	if id == uuid.Nil {
		panic("tenantctx: tenant_id not found in context — middleware misconfiguration")
	}
	return id
}

// TenantSlugFromCtx extracts tenant slug from context
func TenantSlugFromCtx(ctx context.Context) string {
	val := ctx.Value(TenantSlugKey)
	if val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

// TenantPlanFromCtx extracts tenant plan from context
func TenantPlanFromCtx(ctx context.Context) string {
	val := ctx.Value(TenantPlanKey)
	if val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

// CacheKey builds a tenant-scoped cache key to prevent cross-tenant data leaks.
// Format: "t:<tenantID>:<resource>:<id>"
func CacheKey(tenantID uuid.UUID, resource, id string) string {
	return fmt.Sprintf("t:%s:%s:%s", tenantID.String(), resource, id)
}

// Logger returns a slog.Logger with tenant_id automatically injected from context.
// Use this in service layer instead of global.Logger for tenant-aware logging.
func Logger(ctx context.Context) *slog.Logger {
	tenantID := TenantIDFromCtx(ctx)
	slug := TenantSlugFromCtx(ctx)

	attrs := []any{"tenant_id", tenantID.String()}
	if slug != "" {
		attrs = append(attrs, "tenant_slug", slug)
	}
	return slog.With(attrs...)
}
