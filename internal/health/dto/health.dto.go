package dto

import "time"

// SystemHealthResponse is the top-level response for the System Health Dashboard.
type SystemHealthResponse struct {
	QueueHealth   QueueHealth   `json:"queue_health"`
	SLACompliance SLACompliance `json:"sla_compliance"`
	WebhookHealth WebhookHealth `json:"webhook_health"`
	GeneratedAt   time.Time     `json:"generated_at"`
}

// QueueHealth represents the current state of the messaging queue.
type QueueHealth struct {
	UnassignedCount int     `json:"unassigned_count"`
	AssignedCount   int     `json:"assigned_count"`
	OverdueCount    int     `json:"overdue_count"`
	AvgWaitTimeSec  float64 `json:"avg_wait_time_seconds"`
	TotalActive     int     `json:"total_active"`
}

// SLACompliance represents SLA compliance metrics for a given period.
type SLACompliance struct {
	TotalResolved  int     `json:"total_resolved"`
	WithinSLA      int     `json:"within_sla"`
	Breached       int     `json:"breached"`
	ComplianceRate float64 `json:"compliance_rate_percent"`
}

// WebhookHealth represents webhook dispatch reliability metrics.
type WebhookHealth struct {
	TotalDispatched int     `json:"total_dispatched"`
	SuccessCount    int     `json:"success_count"`
	FailedCount     int     `json:"failed_count"`
	PendingRetries  int     `json:"pending_retries"`
	SuccessRate     float64 `json:"success_rate_percent"`
}

// HealthRequest is the request params for the health dashboard endpoint.
type HealthRequest struct {
	EnvironmentID string `form:"environment_id" binding:"required"`
	From          string `form:"from"`
	To            string `form:"to"`
}
