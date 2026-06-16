package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func RateLimitPickup() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:               30,
		Expiration:         1 * time.Minute,
		LimiterMiddleware: limiter.SlidingWindow{},
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"status": "error",
				"error": fiber.Map{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "terlalu banyak request. coba lagi nanti.",
				},
			})
		},
	})
}
