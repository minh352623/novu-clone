package entity

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TenantStatus represents the lifecycle state of a tenant
type TenantStatus string

const (
	TenantPending     TenantStatus = "pending"
	TenantActive      TenantStatus = "active"
	TenantSuspended   TenantStatus = "suspended"
	TenantDeactivated TenantStatus = "deactivated"
)

// validTenantTransitions defines allowed state changes
var validTenantTransitions = map[TenantStatus][]TenantStatus{
	TenantPending:     {TenantActive},
	TenantActive:      {TenantSuspended, TenantDeactivated},
	TenantSuspended:   {TenantActive, TenantDeactivated},
	TenantDeactivated: {TenantActive}, // reactivation within grace period
}

// Tenant represents an organization/workspace in the system
type Tenant struct {
	ID            uuid.UUID    `json:"id"`
	Name          string       `json:"name"`
	Slug          string       `json:"slug"`
	Status        TenantStatus `json:"status"`
	PricingPlanID *uuid.UUID   `json:"pricing_plan_id,omitempty"`
	PlanStartDate time.Time    `json:"plan_start_date"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`

	// Relationships (loaded separately)
	PricingPlan *PricingPlan   `json:"pricing_plan,omitempty"`
	Members     []TenantMember `json:"-"`
}

// Tenant errors
var (
	ErrTenantNameRequired        = errors.New("tenant name is required")
	ErrTenantSlugRequired        = errors.New("tenant slug is required")
	ErrTenantSlugInvalid         = errors.New("tenant slug contains invalid characters")
	ErrTenantSuspended           = errors.New("tenant is suspended")
	ErrTenantPending             = errors.New("tenant is pending approval")
	ErrTenantDeactivated         = errors.New("tenant is deactivated")
	ErrTenantUnavailable         = errors.New("tenant is unavailable")
	ErrInvalidTenantStatusChange = errors.New("invalid tenant status transition")
)

// NewTenant creates a new tenant with active status
func NewTenant(name, slug string) (*Tenant, error) {
	if name == "" {
		return nil, ErrTenantNameRequired
	}
	if slug == "" {
		return nil, ErrTenantSlugRequired
	}

	now := time.Now()
	return &Tenant{
		ID:            uuid.New(),
		Name:          name,
		Slug:          slug,
		Status:        TenantActive,
		PlanStartDate: now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// TransitionTo validates and applies a status change
func (t *Tenant) TransitionTo(target TenantStatus) error {
	allowed, ok := validTenantTransitions[t.Status]
	if !ok {
		return fmt.Errorf("%w: unknown current status %q", ErrInvalidTenantStatusChange, t.Status)
	}
	for _, s := range allowed {
		if s == target {
			t.Status = target
			t.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("%w: %s → %s", ErrInvalidTenantStatusChange, t.Status, target)
}

// IsActive checks if the tenant is in active state
func (t *Tenant) IsActive() bool {
	return t.Status == TenantActive
}

// ValidateStatus returns an appropriate error for non-active tenant states
func (t *Tenant) ValidateStatus() error {
	switch t.Status {
	case TenantActive:
		return nil
	case TenantSuspended:
		return ErrTenantSuspended
	case TenantPending:
		return ErrTenantPending
	case TenantDeactivated:
		return ErrTenantDeactivated
	default:
		return ErrTenantUnavailable
	}
}

// Validate validates the tenant
func (t *Tenant) Validate() error {
	if t.Name == "" {
		return ErrTenantNameRequired
	}
	if t.Slug == "" {
		return ErrTenantSlugRequired
	}
	return nil
}

// AssignPlan assigns a pricing plan to the tenant
func (t *Tenant) AssignPlan(planID uuid.UUID) {
	t.PricingPlanID = &planID
	t.PlanStartDate = time.Now()
	t.UpdatedAt = time.Now()
}

// UpdateName updates the tenant name
func (t *Tenant) UpdateName(name string) error {
	if name == "" {
		return ErrTenantNameRequired
	}
	t.Name = name
	t.UpdatedAt = time.Now()
	return nil
}
