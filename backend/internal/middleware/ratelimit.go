package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter provides per-IP rate limiting keyed by remote address.
type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*limiterEntry
	rps     rate.Limit
	burst   int
}

func NewRateLimiter(rpm int) *RateLimiter {
	if rpm <= 0 {
		rpm = 60
	}
	rl := &RateLimiter{
		entries: make(map[string]*limiterEntry),
		rps:     rate.Limit(float64(rpm) / 60.0),
		burst:   rpm/10 + 1,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	entry, ok := rl.entries[key]
	if !ok {
		entry = &limiterEntry{limiter: rate.NewLimiter(rl.rps, rl.burst)}
		rl.entries[key] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		for key, entry := range rl.entries {
			if time.Since(entry.lastSeen) > 10*time.Minute {
				delete(rl.entries, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if idx := strings.LastIndex(ip, ":"); idx != -1 {
				ip = ip[:idx]
			}
			if !rl.getLimiter(ip).Allow() {
				w.Header().Set("Retry-After", "60")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "rate limit exceeded", "code": "RATE_LIMITED"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
