package ratelimit

import "time"

// RateLimiter checks whether a request identified by key is allowed
// given the configured limit within the specified window.
type RateLimiter interface {
	// Allow returns true if the request is allowed, false if rate limited.
	// limit <= 0 means unlimited (always returns true).
	Allow(key string, limit int, window time.Duration) bool

	// Remaining returns how many requests remain for the given key/window.
	Remaining(key string, limit int, window time.Duration) int
}
