package planlimit_test

import (
	"context"
	"errors"
	"testing"

	"CONVERDA/internal/iam/domain/model/entity"
	"CONVERDA/pkg/planlimit"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// mockCounter implements planlimit.ResourceCounter for testing
type mockCounter struct {
	apps      int
	members   int
	workflows int
	err       error
}

func (m *mockCounter) CountApps(_ context.Context, _ uuid.UUID) (int, error) { return m.apps, m.err }
func (m *mockCounter) CountMembers(_ context.Context, _ uuid.UUID) (int, error) {
	return m.members, m.err
}
func (m *mockCounter) CountWorkflows(_ context.Context, _ uuid.UUID) (int, error) {
	return m.workflows, m.err
}

func TestChecker_CheckLimit(t *testing.T) {
	tenantID := uuid.New()
	ctx := context.Background()

	t.Run("within_limit", func(t *testing.T) {
		counter := &mockCounter{apps: 2}
		checker := planlimit.NewChecker(counter)
		plan := &entity.PricingPlan{MaxApps: 5}

		err := checker.CheckLimit(ctx, plan, tenantID, planlimit.ResourceApps)
		assert.NoError(t, err)
	})

	t.Run("at_limit_returns_error", func(t *testing.T) {
		counter := &mockCounter{members: 5}
		checker := planlimit.NewChecker(counter)
		plan := &entity.PricingPlan{MaxMembers: 5}

		err := checker.CheckLimit(ctx, plan, tenantID, planlimit.ResourceMembers)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, entity.ErrPlanLimitReached))
	})

	t.Run("over_limit_returns_error", func(t *testing.T) {
		counter := &mockCounter{workflows: 15}
		checker := planlimit.NewChecker(counter)
		plan := &entity.PricingPlan{MaxWorkflows: 10}

		err := checker.CheckLimit(ctx, plan, tenantID, planlimit.ResourceWorkflows)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, entity.ErrPlanLimitReached))
	})

	t.Run("unlimited_when_zero", func(t *testing.T) {
		counter := &mockCounter{apps: 999}
		checker := planlimit.NewChecker(counter)
		plan := &entity.PricingPlan{MaxApps: 0} // unlimited

		err := checker.CheckLimit(ctx, plan, tenantID, planlimit.ResourceApps)
		assert.NoError(t, err)
	})

	t.Run("nil_plan_returns_nil", func(t *testing.T) {
		counter := &mockCounter{}
		checker := planlimit.NewChecker(counter)

		err := checker.CheckLimit(ctx, nil, tenantID, planlimit.ResourceApps)
		assert.NoError(t, err)
	})

	t.Run("counter_error_propagates", func(t *testing.T) {
		counter := &mockCounter{err: errors.New("db connection failed")}
		checker := planlimit.NewChecker(counter)
		plan := &entity.PricingPlan{MaxApps: 5}

		err := checker.CheckLimit(ctx, plan, tenantID, planlimit.ResourceApps)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db connection failed")
	})
}
