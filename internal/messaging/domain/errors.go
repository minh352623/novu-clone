package domain

import "errors"

// Sentinel errors for the Messaging module.
// Service layer returns these; controller/error-mapping layer matches with errors.Is().

// Thread status errors
var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrThreadResolved          = errors.New("cannot operate on resolved thread")
)

// Group/Participant errors
var (
	ErrNotGroupThread      = errors.New("not a group thread")
	ErrParticipantNotFound = errors.New("participant not found")
)
