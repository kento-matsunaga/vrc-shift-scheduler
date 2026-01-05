package rest

import (
	"sync"
	"time"
)

// RateLimiter provides in-memory rate limiting using sliding window
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new RateLimiter
// limit: maximum number of requests allowed in the window
// window: time window for rate limiting
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a request from the given key (e.g., IP address) is allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Get existing requests for this key
	requests, exists := rl.requests[key]
	if !exists {
		rl.requests[key] = []time.Time{now}
		return true
	}

	// Filter out old requests (outside the window)
	var validRequests []time.Time
	for _, t := range requests {
		if t.After(cutoff) {
			validRequests = append(validRequests, t)
		}
	}

	// Check if we're under the limit
	if len(validRequests) >= rl.limit {
		rl.requests[key] = validRequests
		return false
	}

	// Add the new request
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests

	return true
}

// cleanup periodically removes old entries from the map
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window)

		for key, requests := range rl.requests {
			var validRequests []time.Time
			for _, t := range requests {
				if t.After(cutoff) {
					validRequests = append(validRequests, t)
				}
			}

			if len(validRequests) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = validRequests
			}
		}
		rl.mu.Unlock()
	}
}

// DefaultClaimRateLimiter creates a rate limiter for the license claim endpoint
// 5 requests per minute per IP
func DefaultClaimRateLimiter() *RateLimiter {
	return NewRateLimiter(5, time.Minute)
}

// DefaultPasswordResetRateLimiter creates a rate limiter for password reset endpoints
// 5 requests per minute per IP - prevents brute force attacks on license keys
func DefaultPasswordResetRateLimiter() *RateLimiter {
	return NewRateLimiter(5, time.Minute)
}
