package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Tenant represents an organization/workspace in the system
type Tenant struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	Slug          string     `json:"slug"`
	PricingPlanID *uuid.UUID `json:"pricing_plan_id,omitempty"`
	PlanStartDate time.Time  `json:"plan_start_date"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	// Relationships (loaded separately)
	PricingPlan *PricingPlan   `json:"pricing_plan,omitempty"`
	Members     []TenantMember `json:"-"`
}

// Tenant errors
var (
	ErrTenantNameRequired = errors.New("tenant name is required")
	ErrTenantSlugRequired = errors.New("tenant slug is required")
	ErrTenantSlugInvalid  = errors.New("tenant slug contains invalid characters")
)

// NewTenant creates a new tenant
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
		PlanStartDate: now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
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
