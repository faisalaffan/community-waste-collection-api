package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
	"github.com/faisalaffan/community-waste-collection-api/internal/service"
)

type mockHouseholdSvc struct {
	createFn  func(req *domain.CreateHouseholdRequest) (*domain.Household, error)
	getByIDFn func(id uuid.UUID) (*domain.Household, error)
	listFn    func(page, perPage int) ([]domain.Household, int64, error)
	deleteFn  func(id uuid.UUID) error
}

func (m *mockHouseholdSvc) Create(req *domain.CreateHouseholdRequest) (*domain.Household, error) {
	return m.createFn(req)
}
func (m *mockHouseholdSvc) GetByID(id uuid.UUID) (*domain.Household, error) {
	return m.getByIDFn(id)
}
func (m *mockHouseholdSvc) List(page, perPage int) ([]domain.Household, int64, error) {
	return m.listFn(page, perPage)
}
func (m *mockHouseholdSvc) Delete(id uuid.UUID) error { return m.deleteFn(id) }

func TestHouseholdHandler_Create_ValidationError(t *testing.T) {
	app := fiber.New()
	h := NewHouseholdHandler(nil)
	app.Post("/households", h.Create)

	body := `{"owner_name":""}`
	req := httptest.NewRequest("POST", "/households", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 422, resp.StatusCode)
}

func TestHouseholdHandler_Create_Success(t *testing.T) {
	app := fiber.New()
	svc := &mockHouseholdSvc{
		createFn: func(req *domain.CreateHouseholdRequest) (*domain.Household, error) {
			return &domain.Household{ID: uuid.New(), OwnerName: req.OwnerName, Address: req.Address}, nil
		},
	}
	h := NewHouseholdHandler(svc)
	app.Post("/households", h.Create)

	body := `{"owner_name":"Budi","address":"Jl. A"}`
	req := httptest.NewRequest("POST", "/households", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 201, resp.StatusCode)
}

func TestHouseholdHandler_Get_InvalidUUID(t *testing.T) {
	app := fiber.New()
	h := NewHouseholdHandler(nil)
	app.Get("/households/:id", h.Get)

	req := httptest.NewRequest("GET", "/households/not-a-uuid", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestHouseholdHandler_Get_NotFound(t *testing.T) {
	app := fiber.New()
	svc := &mockHouseholdSvc{
		getByIDFn: func(id uuid.UUID) (*domain.Household, error) {
			return nil, service.ErrNotFound
		},
	}
	h := NewHouseholdHandler(svc)
	app.Get("/households/:id", h.Get)

	req := httptest.NewRequest("GET", "/households/"+uuid.New().String(), nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 404, resp.StatusCode)
}
