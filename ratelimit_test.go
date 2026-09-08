package westmarches

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiterAllowsWithinBudget(t *testing.T) {
	// 60 req/min => 1 token/sec, burst of 1.
	l := newRateLimiter(60)
	ctx := context.Background()
	// First call should not block (burst token available).
	if err := l.wait(ctx); err != nil {
		t.Fatalf("wait: %v", err)
	}
}

func TestRateLimiterBlocksBeyondBudget(t *testing.T) {
	l := newRateLimiter(60) // 1 token/sec
	ctx := context.Background()
	if err := l.wait(ctx); err != nil {
		t.Fatalf("first wait: %v", err)
	}
	start := time.Now()
	if err := l.wait(ctx); err != nil {
		t.Fatalf("second wait: %v", err)
	}
	if elapsed := time.Since(start); elapsed < 800*time.Millisecond {
		t.Errorf("second wait returned too quickly (%v); expected ~1s refill", elapsed)
	}
}

func TestRateLimiterContextCancellation(t *testing.T) {
	l := newRateLimiter(60) // 1 token/sec
	ctx := context.Background()
	if err := l.wait(ctx); err != nil {
		t.Fatalf("first wait: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := l.wait(ctx); err == nil {
		t.Fatal("expected context error after cancellation")
	}
}

func TestRateLimiterDisabled(t *testing.T) {
	c := NewClient("key", WithRateLimit(0))
	if c.limiter != nil {
		t.Fatal("expected limiter to be nil")
	}
}

func TestRateLimiterZeroValueDefaultsToOne(t *testing.T) {
	l := newRateLimiter(0)
	if l.refillPerSec <= 0 {
		t.Errorf("expected refillPerSec > 0, got %v", l.refillPerSec)
	}
}
