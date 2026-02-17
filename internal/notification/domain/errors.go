package domain

import "errors"

// --- Group Manager Errors ---
var (
	ErrGroupDuplicateKey = errors.New("group with this key already exists")
	ErrGroupNotFound     = errors.New("group not found")
	ErrGroupNotInEnv     = errors.New("group not found in this environment")
)

// --- Layout Manager Errors ---
var (
	ErrLayoutNotFound = errors.New("layout not found")
	ErrLayoutNotInEnv = errors.New("layout not found in this environment")
)

// --- Job Scheduler Errors ---
var (
	ErrJobNotFound         = errors.New("job not found")
	ErrJobNotInEnv         = errors.New("job not found in this environment")
	ErrJobNotCancellable   = errors.New("cannot cancel finished job")
	ErrJobNotPending       = errors.New("job is not pending")
	ErrTemplateCodeMissing = errors.New("template code not found")
)
