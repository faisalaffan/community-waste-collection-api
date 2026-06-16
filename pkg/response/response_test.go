package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestSuccess(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		return Success(c, http.StatusOK, map[string]string{"foo": "bar"})
	})
	resp, _ := app.Test(httptest.NewRequest(http.MethodGet, "/test", http.NoBody))
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	if body["status"] != "success" {
		t.Errorf("status = %v, want success", body["status"])
	}
}

func TestSuccessCreated(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		return SuccessCreated(c, map[string]string{"foo": "bar"})
	})
	resp, _ := app.Test(httptest.NewRequest(http.MethodGet, "/test", http.NoBody))
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
}

func TestSuccessOK(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		return SuccessOK(c, map[string]string{"foo": "bar"})
	})
	resp, _ := app.Test(httptest.NewRequest(http.MethodGet, "/test", http.NoBody))
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		return Error(c, http.StatusNotFound, "NOT_FOUND", "resource not found")
	})
	resp, _ := app.Test(httptest.NewRequest(http.MethodGet, "/test", http.NoBody))
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestValidationError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		return ValidationError(c, "invalid input", []ValidationDetail{
			{Field: "owner_name", Message: "wajib diisi"},
		})
	})
	resp, _ := app.Test(httptest.NewRequest(http.MethodGet, "/test", http.NoBody))
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestSuccessPaginated(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		return SuccessPaginated(c, []string{"a", "b"}, 1, 10, 2)
	})
	resp, _ := app.Test(httptest.NewRequest(http.MethodGet, "/test", http.NoBody))
	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	if body["pagination"] == nil {
		t.Error("pagination should not be nil")
	}
}
