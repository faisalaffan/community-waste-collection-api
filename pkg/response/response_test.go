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

func TestSuccessPaginated_TotalZero(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		return SuccessPaginated(c, []string{}, 1, 10, 0)
	})
	resp, _ := app.Test(httptest.NewRequest(http.MethodGet, "/test", http.NoBody))
	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	pag := body["pagination"].(map[string]interface{})
	if pag["total_pages"].(float64) != 1 {
		t.Errorf("total_pages = %v, want 1", pag["total_pages"])
	}
	if pag["total"].(float64) != 0 {
		t.Errorf("total = %v, want 0", pag["total"])
	}
}

func TestSuccessPaginated_UnevenPages(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		return SuccessPaginated(c, []string{"a", "b", "c"}, 1, 10, 25)
	})
	resp, _ := app.Test(httptest.NewRequest(http.MethodGet, "/test", http.NoBody))
	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	pag := body["pagination"].(map[string]interface{})
	if pag["total_pages"].(float64) != 3 {
		t.Errorf("total_pages = %v, want 3", pag["total_pages"])
	}
	if pag["total"].(float64) != 25 {
		t.Errorf("total = %v, want 25", pag["total"])
	}
}
