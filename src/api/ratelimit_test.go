package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/OpenFactorioServerManager/factorio-server-manager/api"
	"golang.org/x/time/rate"
)

func TestNewRateLimiter(t *testing.T) {
	rl := api.NewRateLimiter(rate.Every(time.Second), 5)
	if rl == nil {
		t.Error("NewRateLimiter returned nil")
	}
}

func TestRateLimiterGetLimiter(t *testing.T) {
	rl := api.NewRateLimiter(rate.Every(time.Second), 5)

	// Get limiter for an IP
	limiter1 := rl.GetLimiter("192.168.1.1")
	if limiter1 == nil {
		t.Error("GetLimiter returned nil for new IP")
	}

	// Get same limiter again
	limiter2 := rl.GetLimiter("192.168.1.1")
	if limiter1 != limiter2 {
		t.Error("GetLimiter should return the same limiter for same IP")
	}

	// Get different limiter for different IP
	limiter3 := rl.GetLimiter("192.168.1.2")
	if limiter1 == limiter3 {
		t.Error("GetLimiter should return different limiters for different IPs")
	}
}

func TestRateLimiterAllowBurst(t *testing.T) {
	rl := api.NewRateLimiter(rate.Every(time.Hour), 3) // Very slow rate, burst of 3

	ip := "10.0.0.1"
	limiter := rl.GetLimiter(ip)

	// Should allow burst number of requests
	for i := 0; i < 3; i++ {
		if !limiter.Allow() {
			t.Errorf("Request %d should have been allowed within burst", i+1)
		}
	}

	// 4th request should be denied (burst exhausted)
	if limiter.Allow() {
		t.Error("Request beyond burst should have been denied")
	}
}

func TestRateLimiterCleanup(t *testing.T) {
	rl := api.NewRateLimiter(rate.Every(time.Second), 5)

	// Add a limiter
	rl.GetLimiter("cleanup-test-ip")

	// Verify it exists
	if rl.GetLimiterCount() != 1 {
		t.Errorf("Expected 1 limiter, got %d", rl.GetLimiterCount())
	}

	// Cleanup with a future cutoff should remove it
	rl.CleanupWithCutoff(time.Now().Add(time.Hour))

	if rl.GetLimiterCount() != 0 {
		t.Errorf("Expected 0 limiters after cleanup, got %d", rl.GetLimiterCount())
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		xForwardedFor  string
		xRealIP        string
		expectedIP     string
	}{
		{
			name:       "RemoteAddr with port",
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "RemoteAddr without port",
			remoteAddr: "192.168.1.1",
			expectedIP: "192.168.1.1",
		},
		{
			name:          "X-Real-IP header",
			remoteAddr:    "127.0.0.1:12345",
			xRealIP:       "10.0.0.1",
			expectedIP:    "10.0.0.1",
		},
		{
			name:          "X-Forwarded-For header",
			remoteAddr:    "127.0.0.1:12345",
			xForwardedFor: "203.0.113.1",
			expectedIP:    "203.0.113.1",
		},
		{
			name:          "X-Forwarded-For takes precedence over X-Real-IP",
			remoteAddr:    "127.0.0.1:12345",
			xForwardedFor: "203.0.113.1",
			xRealIP:       "10.0.0.1",
			expectedIP:    "203.0.113.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			ip := api.GetClientIP(req)
			if ip != tt.expectedIP {
				t.Errorf("GetClientIP() = %q, want %q", ip, tt.expectedIP)
			}
		})
	}
}

func TestLoginRateLimitMiddleware(t *testing.T) {
	// Create a test rate limiter with very restrictive settings
	rl := api.NewRateLimiter(rate.Every(time.Hour), 2) // 2 requests per hour

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with rate limit middleware using the test limiter
	wrapped := api.RateLimitMiddlewareWithLimiter(rl, handler)

	// First 2 requests should succeed (burst)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/api/login", nil)
		req.RemoteAddr = "test-ip:12345"
		rr := httptest.NewRecorder()

		wrapped.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Request %d: expected status %d, got %d", i+1, http.StatusOK, rr.Code)
		}
	}

	// 3rd request should be rate limited
	req := httptest.NewRequest("POST", "/api/login", nil)
	req.RemoteAddr = "test-ip:12345"
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status %d for rate-limited request, got %d", http.StatusTooManyRequests, rr.Code)
	}

	// Check Retry-After header is set
	if rr.Header().Get("Retry-After") == "" {
		t.Error("Expected Retry-After header to be set on rate-limited response")
	}
}
