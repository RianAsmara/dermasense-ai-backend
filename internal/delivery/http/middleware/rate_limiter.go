package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// RateLimiterConfig holds rate limiter configuration.
type RateLimiterConfig struct {
	Max        int
	Window     time.Duration
	KeyPrefix  string
}

// RateLimiter creates a Redis-based rate limiting middleware.
func RateLimiter(redisClient *redis.Client, cfg RateLimiterConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Use user ID if authenticated, otherwise IP
		key := c.IP()
		if userID, ok := c.Locals("userID").(string); ok && userID != "" {
			key = userID
		}

		redisKey := fmt.Sprintf("%s:%s", cfg.KeyPrefix, key)
		ctx := context.Background()

		// Increment request count
		count, err := redisClient.Incr(ctx, redisKey).Result()
		if err != nil {
			// If Redis is down, allow the request
			return c.Next()
		}

		// Set expiry on first request
		if count == 1 {
			redisClient.Expire(ctx, redisKey, cfg.Window)
		}

		if count > int64(cfg.Max) {
			ttl, _ := redisClient.TTL(ctx, redisKey).Result()

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":      "rate limit exceeded",
				"retry_after": ttl.Seconds(),
			})
		}

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.Max))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", int64(cfg.Max)-count))

		return c.Next()
	}
}
