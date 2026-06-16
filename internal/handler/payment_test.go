package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestPaymentHandler_Create_ValidationError(t *testing.T) {
	app := fiber.New()
	h := NewPaymentHandler(nil)
	app.Post("/payments", h.Create)

	body := `{"household_id":""}`
	req := httptest.NewRequest("POST", "/payments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 422, resp.StatusCode)
}

func TestPaymentHandler_Confirm_InvalidUUID(t *testing.T) {
	app := fiber.New()
	h := NewPaymentHandler(nil)
	app.Put("/payments/:id/confirm", h.Confirm)

	req := httptest.NewRequest("PUT", "/payments/not-uuid/confirm", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}
