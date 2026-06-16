package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitPickup_ReturnsNonNil(t *testing.T) {
	handler := RateLimitPickup()
	assert.NotNil(t, handler)
}

func TestRateLimitPickup_Returns429WhenLimitExceeded(t *testing.T) {
	app := fiber.New()
	app.Use(RateLimitPickup())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	for i := 0; i < 31; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		resp, _ := app.Test(req)
		if i < 30 {
			assert.Equal(t, 200, resp.StatusCode, "request %d should be allowed", i+1)
		} else {
			assert.Equal(t, 429, resp.StatusCode, "request %d should be rate limited", i+1)
		}
	}
}
