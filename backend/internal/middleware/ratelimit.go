package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// bucket is an in-memory token bucket for a single key (user or IP).
type bucket struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

func (b *bucket) take() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens = min(b.maxTokens, b.tokens+elapsed*b.refillRate)
	b.lastRefill = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// RateLimiter manages per-key token buckets.
type RateLimiter struct {
	buckets sync.Map
	rpm     int
}

// NewRateLimiter creates a limiter allowing rpm requests per minute per key.
func NewRateLimiter(rpm int) *RateLimiter {
	return &RateLimiter{rpm: rpm}
}

func (rl *RateLimiter) getBucket(key string) *bucket {
	b, _ := rl.buckets.LoadOrStore(key, &bucket{
		tokens:     float64(rl.rpm),
		maxTokens:  float64(rl.rpm),
		refillRate: float64(rl.rpm) / 60.0,
		lastRefill: time.Now(),
	})
	return b.(*bucket)
}

// Limit returns a Gin middleware that rate-limits by the key derived from keyFunc.
func (rl *RateLimiter) Limit(keyFunc func(c *gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		if !rl.getBucket(key).take() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded, please slow down"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// UserKeyFunc extracts the user ID from an authenticated Gin context.
// Falls back to client IP if userID is not set (e.g. on public routes).
func UserKeyFunc(c *gin.Context) string {
	if uid, exists := c.Get("userID"); exists {
		return uid.(uuid.UUID).String()
	}
	return c.ClientIP()
}

// IPKeyFunc uses the client IP address as the rate-limit key.
// Suitable for unauthenticated endpoints like /auth/register and /auth/login.
func IPKeyFunc(c *gin.Context) string {
	return c.ClientIP()
}
