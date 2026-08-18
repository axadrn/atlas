package main

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// rateLimiter is a per-IP token bucket. Buckets refill continuously and idle
// entries are evicted lazily on the next sweep, so memory stays bounded
// without a background goroutine.
type rateLimiter struct {
	mu        sync.Mutex
	buckets   map[string]*bucket
	rate      float64 // tokens per second
	burst     float64
	lastSweep time.Time
	now       func() time.Time
}

type bucket struct {
	tokens float64
	last   time.Time
}

func newRateLimiter(rate, burst float64) *rateLimiter {
	return &rateLimiter{
		buckets: make(map[string]*bucket),
		rate:    rate,
		burst:   burst,
		now:     time.Now,
	}
}

func (l *rateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	if now.Sub(l.lastSweep) > time.Minute {
		l.lastSweep = now
		for key, b := range l.buckets {
			if now.Sub(b.last) > time.Minute {
				delete(l.buckets, key)
			}
		}
	}

	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	}
	b.tokens = min(l.burst, b.tokens+now.Sub(b.last).Seconds()*l.rate)
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// withRateLimit guards a handler with a per-IP limit. The client IP comes
// from CF-Connecting-IP when Cloudflare proxies the request and from the
// socket address otherwise.
func withRateLimit(limiter *rateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("CF-Connecting-IP")
		if ip == "" {
			ip, _, _ = net.SplitHostPort(r.RemoteAddr)
		}
		if !limiter.allow(ip) {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "Too many requests.", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
