package ratelimit

import (
	"CONVERDA/global"
	"sync"
	"time"
)

// bucket tracks request timestamps within a sliding window.
type bucket struct {
	mu         sync.Mutex
	timestamps []time.Time
}

// MemoryRateLimiter is an in-memory sliding window rate limiter.
// Thread-safe, suitable for single-instance deployments.
type MemoryRateLimiter struct {
	buckets sync.Map // map[string]*bucket
	stopCh  chan struct{}
}

// NewMemoryRateLimiter creates a new in-memory rate limiter
// with background cleanup of expired entries every cleanupInterval.
func NewMemoryRateLimiter(cleanupInterval time.Duration) *MemoryRateLimiter {
	rl := &MemoryRateLimiter{
		stopCh: make(chan struct{}),
	}

	// Background cleanup of expired buckets
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		defer func() {
			if r := recover(); r != nil {
				global.Logger.Error("ratelimit: panic recovered in cleanup", "panic", r)
			}
		}()
		for {
			select {
			case <-ticker.C:
				rl.cleanup()
			case <-rl.stopCh:
				return
			}
		}
	}()

	return rl
}

// Stop halts the background cleanup goroutine.
func (rl *MemoryRateLimiter) Stop() {
	close(rl.stopCh)
}

// Allow checks if a request is allowed and records it if so.
func (rl *MemoryRateLimiter) Allow(key string, limit int, window time.Duration) bool {
	if limit <= 0 {
		return true // unlimited
	}

	b := rl.getBucket(key)
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-window)

	// Slide the window: remove expired timestamps
	b.timestamps = filterAfter(b.timestamps, cutoff)

	if len(b.timestamps) >= limit {
		return false
	}

	b.timestamps = append(b.timestamps, now)
	return true
}

// Remaining returns how many requests are left in the current window.
func (rl *MemoryRateLimiter) Remaining(key string, limit int, window time.Duration) int {
	if limit <= 0 {
		return -1 // unlimited
	}

	b := rl.getBucket(key)
	b.mu.Lock()
	defer b.mu.Unlock()

	cutoff := time.Now().Add(-window)
	b.timestamps = filterAfter(b.timestamps, cutoff)

	remaining := limit - len(b.timestamps)
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

func (rl *MemoryRateLimiter) getBucket(key string) *bucket {
	val, _ := rl.buckets.LoadOrStore(key, &bucket{})
	return val.(*bucket)
}

func (rl *MemoryRateLimiter) cleanup() {
	now := time.Now()
	// Remove buckets with no recent activity (older than 1 hour)
	rl.buckets.Range(func(key, value interface{}) bool {
		b := value.(*bucket)
		b.mu.Lock()
		defer b.mu.Unlock()

		if len(b.timestamps) == 0 {
			rl.buckets.Delete(key)
			return true
		}

		// Check if the most recent timestamp is older than 1 hour
		latest := b.timestamps[len(b.timestamps)-1]
		if now.Sub(latest) > 1*time.Hour {
			rl.buckets.Delete(key)
		}
		return true
	})
}

// filterAfter returns only timestamps that are after the cutoff time.
func filterAfter(timestamps []time.Time, cutoff time.Time) []time.Time {
	i := 0
	for _, ts := range timestamps {
		if ts.After(cutoff) {
			timestamps[i] = ts
			i++
		}
	}
	return timestamps[:i]
}
