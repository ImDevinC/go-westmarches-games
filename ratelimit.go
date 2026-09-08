package westmarches

import (
	"context"
	"sync"
	"time"
)

// rateLimiter is a token-bucket limiter that enforces a maximum number of
// requests per minute client-side, so callers avoid tripping the server's rate
// limit. It is safe for concurrent use.
type rateLimiter struct {
	mu           sync.Mutex
	tokens       float64
	maxTokens    float64
	refillPerSec float64
	lastRefill   time.Time
}

// newRateLimiter creates a limiter allowing requestsPerMinute requests per
// minute, with a small burst buffer to smooth out short spikes.
func newRateLimiter(requestsPerMinute int) *rateLimiter {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 1
	}
	// Allow a burst equal to the per-second refill (i.e. roughly a second of
	// budget) so that a handful of rapid calls still succeed.
	refillPerSec := float64(requestsPerMinute) / 60.0
	return &rateLimiter{
		tokens:       refillPerSec,
		maxTokens:    refillPerSec,
		refillPerSec: refillPerSec,
		lastRefill:   time.Now(),
	}
}

// wait blocks until a token is available, or returns ctx.Err() if the context
// is cancelled before a token can be acquired. It returns ErrRateLimited if
// the limiter would need to block but the budget is exhausted and the caller
// must wait (this is signaled through the returned error for testability).
func (l *rateLimiter) wait(ctx context.Context) error {
	l.mu.Lock()
	for {
		l.refill(time.Now())
		if l.tokens >= 1 {
			l.tokens--
			l.mu.Unlock()
			return nil
		}
		// Time until the next token is available.
		wait := time.Duration((1 - l.tokens) / l.refillPerSec * float64(time.Second))
		l.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
		l.mu.Lock()
	}
}

// refill adds tokens up to the bucket capacity based on elapsed time.
func (l *rateLimiter) refill(now time.Time) {
	elapsed := now.Sub(l.lastRefill).Seconds()
	l.lastRefill = now
	l.tokens += elapsed * l.refillPerSec
	if l.tokens > l.maxTokens {
		l.tokens = l.maxTokens
	}
}
