package domain

import "time"

const (
	// Workflow statuses
	WorkflowStatusActive   = "active"
	WorkflowStatusInactive = "inactive"

	// Step Types
	StepTypeChannel = "channel"
	StepTypeDelay   = "delay"
	StepTypeDigest  = "digest"

	// Defaults
	DefaultPollInterval = 30 * time.Second
	DefaultBatchSize    = 50
)
