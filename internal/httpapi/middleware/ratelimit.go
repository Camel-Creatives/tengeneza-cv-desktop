// Package middleware contains request protections backed by Redis in production.
package middleware

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

type rateBucket struct {
	Count   int
	Started time.Time
}
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]rateBucket
	Limit   int
	Window  time.Duration
	Redis   redis.UniversalClient
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{buckets: map[string]rateBucket{}, Limit: limit, Window: window}
}
func NewRedisRateLimiter(client redis.UniversalClient, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{Limit: limit, Window: window, Redis: client}
}
func (r *RateLimiter) Handle(c *fiber.Ctx) error {
	key := c.IP() + ":" + c.Path()
	count := 0
	if r.Redis != nil {
		window := r.Window.Milliseconds()
		if window < 1 {
			window = 1
		}
		result, err := r.Redis.Eval(context.Background(), `local n=redis.call('INCR',KEYS[1]); if n==1 then redis.call('PEXPIRE',KEYS[1],ARGV[1]); end; return n`, []string{"tengenezacv:ratelimit:" + key}, strconv.FormatInt(window, 10)).Int()
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"success": false, "error": fiber.Map{"code": "RATE_LIMIT_STORE_UNAVAILABLE", "message": "Request protection is temporarily unavailable."}})
		}
		count = result
	} else {
		now := time.Now()
		r.mu.Lock()
		bucket := r.buckets[key]
		if bucket.Started.IsZero() || now.Sub(bucket.Started) >= r.Window {
			bucket = rateBucket{Started: now}
		}
		bucket.Count++
		r.buckets[key] = bucket
		count = bucket.Count
		r.mu.Unlock()
	}
	if count > r.Limit {
		retry := int(r.Window.Seconds())
		if retry < 1 {
			retry = 1
		}
		c.Set("Retry-After", strconv.Itoa(retry))
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"success": false, "error": fiber.Map{"code": "RATE_LIMITED", "message": "Too many requests. Try again shortly."}, "meta": fiber.Map{"request_id": c.Locals("request_id"), "timestamp": time.Now().UTC()}})
	}
	return c.Next()
}
