package main

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// rateLimiter implements a per-IP token bucket rate limiter.
// each unique IP gets its own bucket. Tokens refill at a fixed rate,
type rateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*tokenBucket
	rate     time.Duration
	capacity int
}

type tokenBucket struct {
	tokens   int
	lastSeen time.Time
}

// Example: newRateLimiter(10, time.Second) → 10 req/s per IP.
func newRateLimiter(capacity int, rate time.Duration) *rateLimiter {
	rl := &rateLimiter{
		buckets:  make(map[string]*tokenBucket),
		rate:     rate,
		capacity: capacity,
	}

	// evicts idle buckets every minute
	go rl.cleanupLoop()

	return rl
}

// allow returns true if the request from ip should be permitted.
func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, ok := rl.buckets[ip]
	if !ok {
		rl.buckets[ip] = &tokenBucket{tokens: rl.capacity - 1, lastSeen: time.Now()}
		return true
	}

	// refill tokens based on elapsed time
	elapsed := time.Since(bucket.lastSeen)
	refill := int(elapsed / rl.rate)
	if refill > 0 {
		bucket.tokens = min(rl.capacity, bucket.tokens+refill)
		bucket.lastSeen = time.Now()
	}

	if bucket.tokens <= 0 {
		return false
	}

	bucket.tokens--
	return true
}

// cleanupLoop removes buckets that have been idle for more than 5 minutes
// to prevent the map growing unbounded under many unique IPs.
func (rl *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, bucket := range rl.buckets {
			if time.Since(bucket.lastSeen) > 5*time.Minute {
				delete(rl.buckets, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// middleware returns an http.Handler middleware that enforces the rate limit.
func (rl *rateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		if !rl.allow(ip) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate limit exceeded — too many requests"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getClientIP checks for clientIPs supplied in different formatting
func getClientIP(r *http.Request) string {
	// 1. Parse X-Forwarded-For (Render / proxies send: "client, proxy1, proxy2")
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0]) // Extract real client IP
	}

	// parse X-Real-IP
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return strings.TrimSpace(realIP)
	}

	// fallback to RemoteAddr (strip the :port number)
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// min finds the min between two numbers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
