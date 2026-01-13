// Package middleware provides HTTP middleware functions for the API.
package middleware

import (
	"net/http"
	"sync"
	"time"

	"go-api-template/pkg/response"
)

// RateLimiter implements a simple in-memory rate limiter using the token bucket algorithm.
// For production with multiple instances, consider using Redis-based rate limiting.
type RateLimiter struct {
	mu       sync.RWMutex
	clients  map[string]*client
	rate     int           // requests per window
	window   time.Duration // time window
	cleanup  time.Duration // cleanup interval for expired entries
	stopChan chan struct{}
}

type client struct {
	tokens    int
	lastReset time.Time
}

// RateLimitConfig holds the configuration for the rate limiter
type RateLimitConfig struct {
	// Rate is the maximum number of requests allowed per window
	Rate int

	// Window is the time window for rate limiting
	Window time.Duration

	// CleanupInterval is how often to clean up expired client entries
	CleanupInterval time.Duration

	// KeyFunc extracts the rate limit key from the request (default: client IP)
	KeyFunc func(r *http.Request) string
}

// DefaultRateLimitConfig returns a default rate limit configuration.
// 100 requests per minute per IP.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Rate:            100,
		Window:          time.Minute,
		CleanupInterval: 5 * time.Minute,
		KeyFunc:         defaultKeyFunc,
	}
}

// defaultKeyFunc extracts the client IP from the request.
// It checks X-Forwarded-For and X-Real-IP headers first (for reverse proxies).
func defaultKeyFunc(r *http.Request) string {
	// Check for forwarded IP (reverse proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to remote address
	return r.RemoteAddr
}

// NewRateLimiter creates a new rate limiter with the given configuration.
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		clients:  make(map[string]*client),
		rate:     config.Rate,
		window:   config.Window,
		cleanup:  config.CleanupInterval,
		stopChan: make(chan struct{}),
	}

	// Start background cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// cleanupLoop periodically removes expired client entries
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanupExpired()
		case <-rl.stopChan:
			return
		}
	}
}

// cleanupExpired removes client entries that haven't been accessed recently
func (rl *RateLimiter) cleanupExpired() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	threshold := time.Now().Add(-rl.window * 2)
	for key, c := range rl.clients {
		if c.lastReset.Before(threshold) {
			delete(rl.clients, key)
		}
	}
}

// Stop stops the cleanup goroutine. Call this when shutting down.
func (rl *RateLimiter) Stop() {
	close(rl.stopChan)
}

// Allow checks if a request should be allowed based on the rate limit.
// Returns true if allowed, false if rate limited.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	c, exists := rl.clients[key]
	if !exists {
		// New client, create entry with full tokens minus one
		rl.clients[key] = &client{
			tokens:    rl.rate - 1,
			lastReset: now,
		}
		return true
	}

	// Check if window has passed and reset tokens
	if now.Sub(c.lastReset) >= rl.window {
		c.tokens = rl.rate - 1
		c.lastReset = now
		return true
	}

	// Check if tokens available
	if c.tokens > 0 {
		c.tokens--
		return true
	}

	return false
}

// RateLimit returns a middleware that limits requests based on client IP.
func RateLimit(config RateLimitConfig) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(config)

	keyFunc := config.KeyFunc
	if keyFunc == nil {
		keyFunc = defaultKeyFunc
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)

			if !limiter.Allow(key) {
				// Set Retry-After header
				w.Header().Set("Retry-After", "60")
				w.Header().Set("X-RateLimit-Limit", string(rune(config.Rate)))

				response.Error(w, http.StatusTooManyRequests, "Rate limit exceeded. Please try again later.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitWithDefaults returns a rate limiting middleware with default configuration.
func RateLimitWithDefaults() func(http.Handler) http.Handler {
	return RateLimit(DefaultRateLimitConfig())
}
