package main

import (
	"net/http"
	"sync"
	"time"
)

// rateLimiter implements a per-IP token bucket rate limiter.
// Each unique IP gets its own bucket. Tokens refill at a fixed rate,
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
		ip := r.RemoteAddr

		// When behind a proxy or load balancer, prefer the real client IP
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		} else if real := r.Header.Get("X-Real-IP"); real != "" {
			ip = real
		}

		if !rl.allow(ip) {
			http.Error(w, `{"error":"rate limit exceeded — too many requests"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
