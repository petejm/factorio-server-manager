package api

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter stores rate limiters for each IP address
type RateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*visitorLimiter
	rate     rate.Limit
	burst    int
	done     chan struct{}
}

type visitorLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var loginRateLimiter *RateLimiter

// SetupRateLimiter initializes the rate limiter for login attempts
// Allows 5 attempts per minute (1 every 12 seconds on average)
func SetupRateLimiter() {
	loginRateLimiter = NewRateLimiter(rate.Every(12*time.Second), 5)

	// Start cleanup goroutine
	go loginRateLimiter.cleanupLoop()
}

// NewRateLimiter creates a new rate limiter with the given rate and burst
func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*visitorLimiter),
		rate:     r,
		burst:    burst,
		done:     make(chan struct{}),
	}
}

// GetLimiter returns the rate limiter for a given IP, creating one if it doesn't exist
func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.limiters[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = &visitorLimiter{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// cleanupLoop removes stale entries every minute
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-rl.done:
			return
		case <-ticker.C:
			rl.cleanup()
		}
	}
}

// Shutdown stops the cleanup goroutine
func (rl *RateLimiter) Shutdown() {
	close(rl.done)
}

// cleanup removes entries that haven't been seen in over 3 minutes
func (rl *RateLimiter) cleanup() {
	rl.CleanupWithCutoff(time.Now().Add(-3 * time.Minute))
}

// CleanupWithCutoff removes entries that haven't been seen since cutoff time
func (rl *RateLimiter) CleanupWithCutoff(cutoff time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for ip, v := range rl.limiters {
		if v.lastSeen.Before(cutoff) {
			delete(rl.limiters, ip)
		}
	}
}

// GetLimiterCount returns the number of active limiters (for testing)
func (rl *RateLimiter) GetLimiterCount() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return len(rl.limiters)
}

// GetClientIP extracts the client IP from the request
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for reverse proxies)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in the list
		if ip, _, err := net.SplitHostPort(xff); err == nil {
			return ip
		}
		// If no port, use as-is
		return xff
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// LoginRateLimitMiddleware applies rate limiting to login attempts
func LoginRateLimitMiddleware(next http.Handler) http.Handler {
	return RateLimitMiddlewareWithLimiter(loginRateLimiter, next)
}

// RateLimitMiddlewareWithLimiter applies rate limiting with a custom limiter (for testing)
func RateLimitMiddlewareWithLimiter(rl *RateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rl == nil {
			next.ServeHTTP(w, r)
			return
		}

		ip := GetClientIP(r)
		limiter := rl.GetLimiter(ip)

		if !limiter.Allow() {
			log.Printf("Rate limit exceeded for IP: %s on login attempt", ip)
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Too many login attempts. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
