// Package gdpr provides GDPR compliance helpers for data export and erasure.
package gdpr

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DataCategory represents a category of personal data
type DataCategory string

const (
	CategoryProfile       DataCategory = "profile"
	CategoryMessages      DataCategory = "messages"
	CategorySubscriptions DataCategory = "subscriptions"
	CategoryActivity      DataCategory = "activity"
	CategoryAuditLogs     DataCategory = "audit_logs"
)

// ExportRequest represents a GDPR data export request
type ExportRequest struct {
	ID          uuid.UUID      `json:"id"`
	TenantID    uuid.UUID      `json:"tenant_id"`
	UserID      uuid.UUID      `json:"user_id"`
	Categories  []DataCategory `json:"categories"`
	Status      string         `json:"status"` // pending, processing, completed, failed
	RequestedAt time.Time      `json:"requested_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
}

// ErasureRequest represents a GDPR data erasure (right to be forgotten) request
type ErasureRequest struct {
	ID          uuid.UUID      `json:"id"`
	TenantID    uuid.UUID      `json:"tenant_id"`
	UserID      uuid.UUID      `json:"user_id"`
	Categories  []DataCategory `json:"categories"`
	Status      string         `json:"status"` // pending, processing, completed, failed
	RequestedAt time.Time      `json:"requested_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
}

// ExportedData represents a chunk of exported personal data
type ExportedData struct {
	Category DataCategory    `json:"category"`
	Data     json.RawMessage `json:"data"`
}

// DataProvider collects personal data for a specific category
type DataProvider interface {
	Category() DataCategory
	Export(ctx context.Context, tenantID, userID uuid.UUID) (json.RawMessage, error)
	Erase(ctx context.Context, tenantID, userID uuid.UUID) (int64, error) // returns count of erased records
}

// Service orchestrates GDPR data export and erasure
type Service struct {
	providers map[DataCategory]DataProvider
}

// NewService creates a new GDPR service
func NewService(providers ...DataProvider) *Service {
	m := make(map[DataCategory]DataProvider)
	for _, p := range providers {
		m[p.Category()] = p
	}
	return &Service{providers: m}
}

// Export collects all personal data for a user across categories
func (s *Service) Export(ctx context.Context, tenantID, userID uuid.UUID, categories []DataCategory) ([]ExportedData, error) {
	if len(categories) == 0 {
		// Export all categories
		categories = make([]DataCategory, 0, len(s.providers))
		for cat := range s.providers {
			categories = append(categories, cat)
		}
	}

	var result []ExportedData
	for _, cat := range categories {
		provider, ok := s.providers[cat]
		if !ok {
			continue // skip unknown categories
		}
		data, err := provider.Export(ctx, tenantID, userID)
		if err != nil {
			return nil, fmt.Errorf("gdpr: export %s failed: %w", cat, err)
		}
		result = append(result, ExportedData{Category: cat, Data: data})
	}

	return result, nil
}

// Erase removes personal data for a user across categories.
// Returns a summary of erased record counts per category.
func (s *Service) Erase(ctx context.Context, tenantID, userID uuid.UUID, categories []DataCategory) (map[DataCategory]int64, error) {
	if len(categories) == 0 {
		categories = make([]DataCategory, 0, len(s.providers))
		for cat := range s.providers {
			categories = append(categories, cat)
		}
	}

	result := make(map[DataCategory]int64)
	for _, cat := range categories {
		provider, ok := s.providers[cat]
		if !ok {
			continue
		}
		count, err := provider.Erase(ctx, tenantID, userID)
		if err != nil {
			return nil, fmt.Errorf("gdpr: erase %s failed: %w", cat, err)
		}
		result[cat] = count
	}

	return result, nil
}
