package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterBucketsPerKey(t *testing.T) {
	t.Parallel()
	limiter := newRateLimiter(1, 3)
	base := time.Now()
	now := base
	limiter.now = func() time.Time { return now }

	for i := range 3 {
		if !limiter.allow("a") {
			t.Fatalf("request %d within burst should pass", i+1)
		}
	}
	if limiter.allow("a") {
		t.Fatal("request beyond burst should be limited")
	}
	if !limiter.allow("b") {
		t.Fatal("another key must have its own bucket")
	}

	now = base.Add(2 * time.Second)
	if !limiter.allow("a") {
		t.Fatal("bucket should refill over time")
	}
	if !limiter.allow("a") {
		t.Fatal("two seconds at one token per second refill two tokens")
	}
	if limiter.allow("a") {
		t.Fatal("refilled tokens should be spent again")
	}
}

func TestWithRateLimitReturns429(t *testing.T) {
	t.Parallel()
	limiter := newRateLimiter(1, 2)
	handler := withRateLimit(limiter, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	var last *httptest.ResponseRecorder
	for range 3 {
		last = httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/fragments/place-search?q=x", nil)
		request.RemoteAddr = "203.0.113.7:1234"
		handler.ServeHTTP(last, request)
	}
	if last.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after burst, got %d", last.Code)
	}
	if last.Header().Get("Retry-After") == "" {
		t.Fatal("expected a Retry-After header")
	}
}
