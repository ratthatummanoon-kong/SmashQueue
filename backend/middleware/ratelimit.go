package middleware

import (
	"backend/model"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter holds rate limiters per IP
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    burst,
	}

	// Cleanup old entries periodically
	go rl.cleanup()

	return rl
}

// getLimiter returns the rate limiter for a given IP
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = limiter
	}

	return limiter
}

// cleanup removes old limiters periodically
func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(10 * time.Minute)
		rl.mu.Lock()
		// Simple cleanup - in production, track last access time
		if len(rl.limiters) > 10000 {
			rl.limiters = make(map[string]*rate.Limiter)
		}
		rl.mu.Unlock()
	}
}

// RateLimit creates a rate limiting middleware
// Uses 10 requests per second with burst of 20 by default
func RateLimit(limiter *RateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getIP(r)
			l := limiter.getLimiter(ip)

			if !l.Allow() {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(model.NewErrorResponse(
					http.StatusTooManyRequests,
					"Rate limit exceeded",
					"Please wait before making more requests",
				))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// StrictRateLimit is a stricter rate limit for auth endpoints
// 10 requests per minute with burst of 5
func StrictRateLimit() *RateLimiter {
	return NewRateLimiter(rate.Every(6*time.Second), 5)
}

// DefaultRateLimit for general API endpoints
// 100 requests per minute with burst of 20
func DefaultRateLimit() *RateLimiter {
	return NewRateLimiter(rate.Every(600*time.Millisecond), 20)
}

// getIP extracts the client IP from the request
func getIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
