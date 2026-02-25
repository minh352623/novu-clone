// Package planlimit provides plan-based limit enforcement for multi-tenant operations.
package planlimit

import (
	"context"
	"fmt"

	"CONVERDA/internal/iam/domain/model/entity"

	"github.com/google/uuid"
)

// Resource represents a countable resource type
type Resource string

const (
	ResourceApps      Resource = "apps"
	ResourceMembers   Resource = "members"
	ResourceWorkflows Resource = "workflows"
)

// ResourceCounter retrieves current resource counts for a tenant
type ResourceCounter interface {
	CountApps(ctx context.Context, tenantID uuid.UUID) (int, error)
	CountMembers(ctx context.Context, tenantID uuid.UUID) (int, error)
	CountWorkflows(ctx context.Context, tenantID uuid.UUID) (int, error)
}

// Checker validates whether a tenant can create new resources based on their plan
type Checker struct {
	counter ResourceCounter
}

// NewChecker creates a new plan limit checker
func NewChecker(counter ResourceCounter) *Checker {
	return &Checker{counter: counter}
}

// CheckLimit verifies the tenant has not exceeded the plan limit for the given resource.
// Returns nil if OK, ErrPlanLimitReached if limit reached.
func (c *Checker) CheckLimit(ctx context.Context, plan *entity.PricingPlan, tenantID uuid.UUID, resource Resource) error {
	if plan == nil {
		return nil // no plan = no limits (backward compatibility)
	}

	var limit int
	var current int
	var err error

	switch resource {
	case ResourceApps:
		limit = plan.MaxApps
		current, err = c.counter.CountApps(ctx, tenantID)
	case ResourceMembers:
		limit = plan.MaxMembers
		current, err = c.counter.CountMembers(ctx, tenantID)
	case ResourceWorkflows:
		limit = plan.MaxWorkflows
		current, err = c.counter.CountWorkflows(ctx, tenantID)
	default:
		return fmt.Errorf("unknown resource type: %s", resource)
	}

	if err != nil {
		return fmt.Errorf("failed to count %s: %w", resource, err)
	}

	if !plan.IsWithinLimit(limit, current) {
		return fmt.Errorf("%w: %s limit of %d reached (current: %d)", entity.ErrPlanLimitReached, resource, limit, current)
	}

	return nil
}
