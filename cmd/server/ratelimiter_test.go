package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// okHandler is a trivial next handler used to verify pass-through behaviour
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

// TestRateLimiter_Allow_FirstRequest verifies a brand-new IP is always allowed.
func TestRateLimiter_Allow_FirstRequest(t *testing.T) {
	rl := &rateLimiter{
		buckets:  make(map[string]*tokenBucket),
		rate:     time.Second,
		capacity: 5,
	}

	assert.True(t, rl.allow("1.2.3.4"), "first request from a new IP should always be allowed")
}

// TestRateLimiter_Allow_ExhaustsCapacity verifies requests are blocked once the bucket is empty.
func TestRateLimiter_Allow_ExhaustsCapacity(t *testing.T) {
	capacity := 3

	rl := &rateLimiter{
		buckets:  make(map[string]*tokenBucket),
		rate:     time.Hour, // very long refill so tokens never refill during the test
		capacity: capacity,
	}

	ip := "10.0.0.1"

	// First `capacity` requests should all pass
	for i := 0; i < capacity; i++ {
		assert.True(t, rl.allow(ip), "request %d should be allowed", i+1)
	}

	// The next request should be blocked
	assert.False(t, rl.allow(ip), "request after capacity is exhausted should be denied")
}

// TestRateLimiter_Allow_RefillsOverTime verifies tokens are restored after the rate window passes.
func TestRateLimiter_Allow_RefillsOverTime(t *testing.T) {
	rl := &rateLimiter{
		buckets:  make(map[string]*tokenBucket),
		rate:     50 * time.Millisecond,
		capacity: 2,
	}

	ip := "192.168.1.1"

	// Drain the bucket
	rl.allow(ip)
	rl.allow(ip)
	assert.False(t, rl.allow(ip), "bucket should be empty after capacity is drained")

	// Wait for tokens to refill
	time.Sleep(120 * time.Millisecond)

	assert.True(t, rl.allow(ip), "request should be allowed after tokens have refilled")
}

// TestRateLimiter_Allow_IsolatesPerIP verifies different IPs have independent buckets.
func TestRateLimiter_Allow_IsolatesPerIP(t *testing.T) {
	rl := &rateLimiter{
		buckets:  make(map[string]*tokenBucket),
		rate:     time.Hour,
		capacity: 1,
	}

	ip1 := "10.0.0.1"
	ip2 := "10.0.0.2"

	// Exhaust ip1's bucket
	rl.allow(ip1)
	assert.False(t, rl.allow(ip1), "ip1 should be rate limited")

	// ip2 should be completely unaffected
	assert.True(t, rl.allow(ip2), "ip2 should have its own independent bucket")
}

// TestRateLimiter_Middleware_Passes verifies allowed requests reach the next handler.
func TestRateLimiter_Middleware_Passes(t *testing.T) {
	rl := &rateLimiter{
		buckets:  make(map[string]*tokenBucket),
		rate:     time.Second,
		capacity: 10,
	}

	handler := rl.middleware(okHandler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "5.5.5.5:1234"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestRateLimiter_Middleware_Blocks verifies exhausted buckets return 429.
func TestRateLimiter_Middleware_Blocks(t *testing.T) {
	rl := &rateLimiter{
		buckets:  make(map[string]*tokenBucket),
		rate:     time.Hour,
		capacity: 1,
	}

	handler := rl.middleware(okHandler)
	ip := "7.7.7.7"

	// First request — should pass
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = ip + ":9999"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Second request — bucket empty, should be blocked
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = ip + ":9999"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusTooManyRequests, rec2.Code)
}

// TestRateLimiter_Middleware_RespectsXForwardedFor verifies proxy headers are preferred over RemoteAddr.
func TestRateLimiter_Middleware_RespectsXForwardedFor(t *testing.T) {
	rl := &rateLimiter{
		buckets:  make(map[string]*tokenBucket),
		rate:     time.Hour,
		capacity: 1,
	}

	handler := rl.middleware(okHandler)

	// Two requests with the same X-Forwarded-For but different RemoteAddr (as a proxy would send)
	makeReq := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "proxy:80"
		req.Header.Set("X-Forwarded-For", "7.7.7.7")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	assert.Equal(t, http.StatusOK, makeReq().Code, "first request via proxy should pass")
	assert.Equal(t, http.StatusTooManyRequests, makeReq().Code, "second request with same forwarded IP should be blocked")
}

// TestRateLimiter_Middleware_RespectsXRealIP verifies X-Real-IP is used when X-Forwarded-For is absent.
func TestRateLimiter_Middleware_RespectsXRealIP(t *testing.T) {
	rl := &rateLimiter{
		buckets:  make(map[string]*tokenBucket),
		rate:     time.Hour,
		capacity: 1,
	}

	handler := rl.middleware(okHandler)

	makeReq := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "proxy:80"
		req.Header.Set("X-Real-IP", "8.8.8.8")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	assert.Equal(t, http.StatusOK, makeReq().Code, "first request should pass")
	assert.Equal(t, http.StatusTooManyRequests, makeReq().Code, "second request with same X-Real-IP should be blocked")
}
