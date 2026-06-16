package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPickupHandler_Create_ValidationError_EmptyType(t *testing.T) {
	app := fiber.New()
	h := NewPickupHandler(nil)
	app.Post("/pickups", h.Create)

	body := `{"household_id":"` + uuid.New().String() + `","type":""}`
	req := httptest.NewRequest("POST", "/pickups", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 422, resp.StatusCode)
}

func TestPickupHandler_Create_InvalidType(t *testing.T) {
	app := fiber.New()
	h := NewPickupHandler(nil)
	app.Post("/pickups", h.Create)

	body := `{"household_id":"` + uuid.New().String() + `","type":"metal"}`
	req := httptest.NewRequest("POST", "/pickups", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 422, resp.StatusCode)
}
