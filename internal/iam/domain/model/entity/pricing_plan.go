package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// PricingPlan represents a subscription plan for tenants
type PricingPlan struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	MonthlyCredits int64     `json:"monthly_credits"`
	Price          float64   `json:"price"`
	Currency       string    `json:"currency"`
	Description    *string   `json:"description,omitempty"`
	IsActive       bool      `json:"is_active"`
	IsDefault      bool      `json:"is_default"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// PricingPlan errors
var (
	ErrPlanNameRequired = errors.New("plan name is required")
	ErrPlanSlugRequired = errors.New("plan slug is required")
	ErrPlanNotActive    = errors.New("pricing plan is not active")
)

// NewPricingPlan creates a new pricing plan
func NewPricingPlan(name, slug string, monthlyCredits int64, price float64) (*PricingPlan, error) {
	if name == "" {
		return nil, ErrPlanNameRequired
	}
	if slug == "" {
		return nil, ErrPlanSlugRequired
	}

	now := time.Now()
	return &PricingPlan{
		ID:             uuid.New(),
		Name:           name,
		Slug:           slug,
		MonthlyCredits: monthlyCredits,
		Price:          price,
		Currency:       "USD",
		IsActive:       true,
		IsDefault:      false,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// Validate validates the pricing plan
func (p *PricingPlan) Validate() error {
	if p.Name == "" {
		return ErrPlanNameRequired
	}
	if p.Slug == "" {
		return ErrPlanSlugRequired
	}
	return nil
}

// Activate activates the plan
func (p *PricingPlan) Activate() {
	p.IsActive = true
	p.UpdatedAt = time.Now()
}

// Deactivate deactivates the plan
func (p *PricingPlan) Deactivate() {
	p.IsActive = false
	p.UpdatedAt = time.Now()
}

// SetAsDefault sets this plan as the default
func (p *PricingPlan) SetAsDefault() {
	p.IsDefault = true
	p.UpdatedAt = time.Now()
}
