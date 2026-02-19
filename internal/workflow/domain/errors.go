package domain

import "errors"

var (
	ErrWorkflowNotFound         = errors.New("workflow not found")
	ErrWorkflowActive           = errors.New("workflow is active")
	ErrStepNotFound             = errors.New("workflow step not found")
	ErrTriggerIdentifierMissing = errors.New("trigger identifier is missing")
)
