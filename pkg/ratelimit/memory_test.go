package ratelimit_test

import (
	"sync"
	"testing"
	"time"

	"CONVERDA/pkg/ratelimit"

	"github.com/stretchr/testify/assert"
)

func TestMemoryRateLimiter_BasicAllow(t *testing.T) {
	rl := ratelimit.NewMemoryRateLimiter(1 * time.Minute)
	defer rl.Stop()

	// Allow 3 requests per minute
	assert.True(t, rl.Allow("test-key", 3, 1*time.Minute))
	assert.True(t, rl.Allow("test-key", 3, 1*time.Minute))
	assert.True(t, rl.Allow("test-key", 3, 1*time.Minute))

	// 4th should be denied
	assert.False(t, rl.Allow("test-key", 3, 1*time.Minute))
}

func TestMemoryRateLimiter_UnlimitedWhenZero(t *testing.T) {
	rl := ratelimit.NewMemoryRateLimiter(1 * time.Minute)
	defer rl.Stop()

	// limit=0 => always allowed
	for i := 0; i < 100; i++ {
		assert.True(t, rl.Allow("unlimited", 0, 1*time.Minute))
	}
}

func TestMemoryRateLimiter_NegativeLimitMeansUnlimited(t *testing.T) {
	rl := ratelimit.NewMemoryRateLimiter(1 * time.Minute)
	defer rl.Stop()

	for i := 0; i < 50; i++ {
		assert.True(t, rl.Allow("neg", -1, 1*time.Minute))
	}
}

func TestMemoryRateLimiter_SeparateKeys(t *testing.T) {
	rl := ratelimit.NewMemoryRateLimiter(1 * time.Minute)
	defer rl.Stop()

	// Exhaust key-a
	rl.Allow("key-a", 1, 1*time.Minute)
	assert.False(t, rl.Allow("key-a", 1, 1*time.Minute))

	// key-b should still work
	assert.True(t, rl.Allow("key-b", 1, 1*time.Minute))
}

func TestMemoryRateLimiter_WindowExpiry(t *testing.T) {
	rl := ratelimit.NewMemoryRateLimiter(1 * time.Minute)
	defer rl.Stop()

	// Allow 2 requests per 50ms
	assert.True(t, rl.Allow("expiry", 2, 50*time.Millisecond))
	assert.True(t, rl.Allow("expiry", 2, 50*time.Millisecond))
	assert.False(t, rl.Allow("expiry", 2, 50*time.Millisecond))

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	// Should be allowed again
	assert.True(t, rl.Allow("expiry", 2, 50*time.Millisecond))
}

func TestMemoryRateLimiter_Remaining(t *testing.T) {
	rl := ratelimit.NewMemoryRateLimiter(1 * time.Minute)
	defer rl.Stop()

	assert.Equal(t, 5, rl.Remaining("rem", 5, 1*time.Minute))

	rl.Allow("rem", 5, 1*time.Minute)
	assert.Equal(t, 4, rl.Remaining("rem", 5, 1*time.Minute))

	rl.Allow("rem", 5, 1*time.Minute)
	rl.Allow("rem", 5, 1*time.Minute)
	assert.Equal(t, 2, rl.Remaining("rem", 5, 1*time.Minute))

	// Unlimited returns -1
	assert.Equal(t, -1, rl.Remaining("x", 0, 1*time.Minute))
}

func TestMemoryRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := ratelimit.NewMemoryRateLimiter(1 * time.Minute)
	defer rl.Stop()

	limit := 100
	var wg sync.WaitGroup
	allowed := make(chan bool, 200)

	// Fire 200 concurrent requests with limit 100
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed <- rl.Allow("concurrent", limit, 1*time.Minute)
		}()
	}

	wg.Wait()
	close(allowed)

	allowedCount := 0
	deniedCount := 0
	for a := range allowed {
		if a {
			allowedCount++
		} else {
			deniedCount++
		}
	}

	assert.Equal(t, 100, allowedCount, "exactly 100 requests should be allowed")
	assert.Equal(t, 100, deniedCount, "exactly 100 requests should be denied")
}
