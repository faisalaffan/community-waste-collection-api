package handler

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
	"github.com/faisalaffan/community-waste-collection-api/internal/repository"
	"github.com/faisalaffan/community-waste-collection-api/internal/service"
)

type mockPickupSvc struct {
	createFn   func(req *domain.CreatePickupRequest) (*domain.WastePickup, error)
	getByIDFn  func(id uuid.UUID) (*domain.WastePickup, error)
	listFn     func(filter repository.PickupFilter) ([]domain.WastePickup, int64, error)
	updateFn   func(id uuid.UUID, req *domain.CreatePickupRequest) (*domain.WastePickup, error)
	deleteFn   func(id uuid.UUID) error
	scheduleFn func(id uuid.UUID, req *domain.SchedulePickupRequest) (*domain.WastePickup, error)
	completeFn func(id uuid.UUID) (*domain.WastePickup, *domain.Payment, error)
	cancelFn   func(id uuid.UUID) (*domain.WastePickup, error)
}

func (m *mockPickupSvc) Create(req *domain.CreatePickupRequest) (*domain.WastePickup, error) {
	return m.createFn(req)
}
func (m *mockPickupSvc) GetByID(id uuid.UUID) (*domain.WastePickup, error) {
	return m.getByIDFn(id)
}
func (m *mockPickupSvc) List(filter repository.PickupFilter) ([]domain.WastePickup, int64, error) {
	return m.listFn(filter)
}
func (m *mockPickupSvc) Update(id uuid.UUID, req *domain.CreatePickupRequest) (*domain.WastePickup, error) {
	return m.updateFn(id, req)
}
func (m *mockPickupSvc) Delete(id uuid.UUID) error {
	return m.deleteFn(id)
}
func (m *mockPickupSvc) Schedule(id uuid.UUID, req *domain.SchedulePickupRequest) (*domain.WastePickup, error) {
	return m.scheduleFn(id, req)
}
func (m *mockPickupSvc) Complete(id uuid.UUID) (*domain.WastePickup, *domain.Payment, error) {
	return m.completeFn(id)
}
func (m *mockPickupSvc) Cancel(id uuid.UUID) (*domain.WastePickup, error) {
	return m.cancelFn(id)
}

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

func TestPickupHandler_Create_Success(t *testing.T) {
	app := fiber.New()
	svc := &mockPickupSvc{
		createFn: func(req *domain.CreatePickupRequest) (*domain.WastePickup, error) {
			return &domain.WastePickup{
				ID:          uuid.New(),
				HouseholdID: req.HouseholdID,
				Type:        req.Type,
				Status:      domain.PickupStatusPending,
			}, nil
		},
	}
	h := NewPickupHandler(svc)
	app.Post("/pickups", h.Create)

	body := `{"household_id":"` + uuid.New().String() + `","type":"organic"}`
	req := httptest.NewRequest("POST", "/pickups", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 201, resp.StatusCode)
}

func TestPickupHandler_Create_BR01_Blocked(t *testing.T) {
	app := fiber.New()
	svc := &mockPickupSvc{
		createFn: func(req *domain.CreatePickupRequest) (*domain.WastePickup, error) {
			return nil, service.ErrPendingPayment
		},
	}
	h := NewPickupHandler(svc)
	app.Post("/pickups", h.Create)

	body := `{"household_id":"` + uuid.New().String() + `","type":"organic"}`
	req := httptest.NewRequest("POST", "/pickups", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 409, resp.StatusCode)
}

func TestPickupHandler_List_SuccessWithFilters(t *testing.T) {
	app := fiber.New()
	svc := &mockPickupSvc{
		listFn: func(filter repository.PickupFilter) ([]domain.WastePickup, int64, error) {
			assert.NotEqual(t, uuid.Nil, filter.HouseholdID)
			assert.Equal(t, "pending", filter.Status)
			return []domain.WastePickup{
				{ID: uuid.New(), HouseholdID: filter.HouseholdID, Type: domain.PickupTypeOrganic, Status: domain.PickupStatusPending},
			}, 1, nil
		},
	}
	h := NewPickupHandler(svc)
	app.Get("/pickups", h.List)

	req := httptest.NewRequest("GET", "/pickups?household_id="+uuid.New().String()+"&status=pending", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestPickupHandler_List_Empty(t *testing.T) {
	app := fiber.New()
	svc := &mockPickupSvc{
		listFn: func(filter repository.PickupFilter) ([]domain.WastePickup, int64, error) {
			return []domain.WastePickup{}, 0, nil
		},
	}
	h := NewPickupHandler(svc)
	app.Get("/pickups", h.List)

	req := httptest.NewRequest("GET", "/pickups", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestPickupHandler_Schedule_Success(t *testing.T) {
	app := fiber.New()
	svc := &mockPickupSvc{
		scheduleFn: func(id uuid.UUID, req *domain.SchedulePickupRequest) (*domain.WastePickup, error) {
			return &domain.WastePickup{
				ID: id, Status: domain.PickupStatusScheduled, PickupDate: &req.PickupDate,
			}, nil
		},
	}
	h := NewPickupHandler(svc)
	app.Put("/pickups/:id/schedule", h.Schedule)

	body := `{"pickup_date":"2026-06-20T10:00:00Z"}`
	req := httptest.NewRequest("PUT", "/pickups/"+uuid.New().String()+"/schedule", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestPickupHandler_Schedule_InvalidUUID(t *testing.T) {
	app := fiber.New()
	h := NewPickupHandler(nil)
	app.Put("/pickups/:id/schedule", h.Schedule)

	req := httptest.NewRequest("PUT", "/pickups/not-uuid/schedule", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestPickupHandler_Schedule_NotFound(t *testing.T) {
	app := fiber.New()
	svc := &mockPickupSvc{
		scheduleFn: func(id uuid.UUID, req *domain.SchedulePickupRequest) (*domain.WastePickup, error) {
			return nil, service.ErrNotFound
		},
	}
	h := NewPickupHandler(svc)
	app.Put("/pickups/:id/schedule", h.Schedule)

	body := `{"pickup_date":"2026-06-20T10:00:00Z"}`
	req := httptest.NewRequest("PUT", "/pickups/"+uuid.New().String()+"/schedule", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestPickupHandler_Complete_Success(t *testing.T) {
	app := fiber.New()
	pickup := &domain.WastePickup{ID: uuid.New(), Status: domain.PickupStatusCompleted}
	payment := &domain.Payment{ID: uuid.New(), Amount: 50000}
	svc := &mockPickupSvc{
		completeFn: func(id uuid.UUID) (*domain.WastePickup, *domain.Payment, error) {
			return pickup, payment, nil
		},
	}
	h := NewPickupHandler(svc)
	app.Put("/pickups/:id/complete", h.Complete)

	req := httptest.NewRequest("PUT", "/pickups/"+uuid.New().String()+"/complete", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestPickupHandler_Complete_NotFound(t *testing.T) {
	app := fiber.New()
	svc := &mockPickupSvc{
		completeFn: func(id uuid.UUID) (*domain.WastePickup, *domain.Payment, error) {
			return nil, nil, service.ErrNotFound
		},
	}
	h := NewPickupHandler(svc)
	app.Put("/pickups/:id/complete", h.Complete)

	req := httptest.NewRequest("PUT", "/pickups/"+uuid.New().String()+"/complete", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestPickupHandler_Cancel_Success(t *testing.T) {
	app := fiber.New()
	svc := &mockPickupSvc{
		cancelFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{ID: id, Status: domain.PickupStatusCanceled}, nil
		},
	}
	h := NewPickupHandler(svc)
	app.Put("/pickups/:id/cancel", h.Cancel)

	req := httptest.NewRequest("PUT", "/pickups/"+uuid.New().String()+"/cancel", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestPickupHandler_Cancel_InvalidUUID(t *testing.T) {
	app := fiber.New()
	h := NewPickupHandler(nil)
	app.Put("/pickups/:id/cancel", h.Cancel)

	req := httptest.NewRequest("PUT", "/pickups/not-uuid/cancel", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestPickupHandler_Cancel_NotFound(t *testing.T) {
	app := fiber.New()
	svc := &mockPickupSvc{
		cancelFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return nil, service.ErrNotFound
		},
	}
	h := NewPickupHandler(svc)
	app.Put("/pickups/:id/cancel", h.Cancel)

	req := httptest.NewRequest("PUT", "/pickups/"+uuid.New().String()+"/cancel", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestPickupHandler_Cancel_InvalidStatus(t *testing.T) {
	app := fiber.New()
	svc := &mockPickupSvc{
		cancelFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return nil, errors.New("hanya pickup dengan status pending yang dapat dibatalkan")
		},
	}
	h := NewPickupHandler(svc)
	app.Put("/pickups/:id/cancel", h.Cancel)

	// The mock returns a generic error (not ErrPickupNotPending), so handler returns 500
	// We need to test the ErrPickupNotPending path separately
	req := httptest.NewRequest("PUT", "/pickups/"+uuid.New().String()+"/cancel", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 500, resp.StatusCode)
}

func TestPickupHandler_Cancel_InvalidStatus_409(t *testing.T) {
	app := fiber.New()
	svc := &mockPickupSvc{
		cancelFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return nil, service.ErrPickupNotPending
		},
	}
	h := NewPickupHandler(svc)
	app.Put("/pickups/:id/cancel", h.Cancel)

	req := httptest.NewRequest("PUT", "/pickups/"+uuid.New().String()+"/cancel", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 409, resp.StatusCode)
}

func TestPickupHandler_Update_InvalidID(t *testing.T) {
	app := fiber.New()
	h := NewPickupHandler(nil)
	app.Put("/pickups/:id", h.Update)
	req := httptest.NewRequest("PUT", "/pickups/not-uuid", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestPickupHandler_Update_InvalidBody(t *testing.T) {
	app := fiber.New()
	h := NewPickupHandler(nil)
	app.Put("/pickups/:id", h.Update)
	req := httptest.NewRequest("PUT", "/pickups/"+uuid.New().String(), strings.NewReader("bad json"))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 422, resp.StatusCode)
}

func TestPickupHandler_Update_NotFound(t *testing.T) {
	svc := &mockPickupSvc{
		updateFn: func(id uuid.UUID, req *domain.CreatePickupRequest) (*domain.WastePickup, error) {
			return nil, service.ErrNotFound
		},
	}
	app := fiber.New()
	h := NewPickupHandler(svc)
	app.Put("/pickups/:id", h.Update)
	req := httptest.NewRequest("PUT", "/pickups/"+uuid.New().String(), strings.NewReader(`{"type":"organic"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestPickupHandler_Update_Success(t *testing.T) {
	id := uuid.New()
	svc := &mockPickupSvc{
		updateFn: func(uid uuid.UUID, req *domain.CreatePickupRequest) (*domain.WastePickup, error) {
			return &domain.WastePickup{ID: uid, Type: req.Type}, nil
		},
	}
	app := fiber.New()
	h := NewPickupHandler(svc)
	app.Put("/pickups/:id", h.Update)
	req := httptest.NewRequest("PUT", "/pickups/"+id.String(), strings.NewReader(`{"type":"plastic"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestPickupHandler_Update_InvalidType(t *testing.T) {
	svc := &mockPickupSvc{
		updateFn: func(id uuid.UUID, req *domain.CreatePickupRequest) (*domain.WastePickup, error) {
			return nil, service.ErrInvalidPickupType
		},
	}
	app := fiber.New()
	h := NewPickupHandler(svc)
	app.Put("/pickups/:id", h.Update)
	req := httptest.NewRequest("PUT", "/pickups/"+uuid.New().String(), strings.NewReader(`{"type":"invalid"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 422, resp.StatusCode)
}

func TestPickupHandler_Delete_InvalidID(t *testing.T) {
	app := fiber.New()
	h := NewPickupHandler(nil)
	app.Delete("/pickups/:id", h.Delete)
	req := httptest.NewRequest("DELETE", "/pickups/not-uuid", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestPickupHandler_Delete_NotFound(t *testing.T) {
	svc := &mockPickupSvc{
		deleteFn: func(id uuid.UUID) error { return service.ErrNotFound },
	}
	app := fiber.New()
	h := NewPickupHandler(svc)
	app.Delete("/pickups/:id", h.Delete)
	req := httptest.NewRequest("DELETE", "/pickups/"+uuid.New().String(), nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestPickupHandler_Delete_Success(t *testing.T) {
	svc := &mockPickupSvc{
		deleteFn: func(id uuid.UUID) error { return nil },
	}
	app := fiber.New()
	h := NewPickupHandler(svc)
	app.Delete("/pickups/:id", h.Delete)
	req := httptest.NewRequest("DELETE", "/pickups/"+uuid.New().String(), nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestPickupHandler_Update_InvalidStatus(t *testing.T) {
	svc := &mockPickupSvc{
		updateFn: func(id uuid.UUID, req *domain.CreatePickupRequest) (*domain.WastePickup, error) {
			return nil, service.ErrPickupNotPending
		},
	}
	app := fiber.New()
	h := NewPickupHandler(svc)
	app.Put("/pickups/:id", h.Update)
	req := httptest.NewRequest("PUT", "/pickups/"+uuid.New().String(), strings.NewReader(`{"type":"organic"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 409, resp.StatusCode)
}

func TestPickupHandler_Delete_InvalidStatus(t *testing.T) {
	svc := &mockPickupSvc{
		deleteFn: func(id uuid.UUID) error { return service.ErrPickupNotPending },
	}
	app := fiber.New()
	h := NewPickupHandler(svc)
	app.Delete("/pickups/:id", h.Delete)
	req := httptest.NewRequest("DELETE", "/pickups/"+uuid.New().String(), nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 409, resp.StatusCode)
}
func TestPickupHandler_Create_InvalidJSON(t *testing.T) {
	app := fiber.New()
	h := NewPickupHandler(nil)
	app.Post("/pickups", h.Create)

	body := `{not-json}`
	req := httptest.NewRequest("POST", "/pickups", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 422, resp.StatusCode)
}

func TestPickupHandler_Create_EmptyHouseholdID(t *testing.T) {
	app := fiber.New()
	h := NewPickupHandler(nil)
	app.Post("/pickups", h.Create)

	body := `{"type":"organic","household_id":"00000000-0000-0000-0000-000000000000"}`
	req := httptest.NewRequest("POST", "/pickups", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 422, resp.StatusCode)
}

func TestPickupHandler_Create_InternalError(t *testing.T) {
	app := fiber.New()
	svc := &mockPickupSvc{
		createFn: func(req *domain.CreatePickupRequest) (*domain.WastePickup, error) {
			return nil, errors.New("db error")
		},
	}
	h := NewPickupHandler(svc)
	app.Post("/pickups", h.Create)

	body := `{"household_id":"` + uuid.New().String() + `","type":"organic"}`
	req := httptest.NewRequest("POST", "/pickups", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 500, resp.StatusCode)
}

func TestPickupHandler_List_InternalError(t *testing.T) {
	app := fiber.New()
	svc := &mockPickupSvc{
		listFn: func(filter repository.PickupFilter) ([]domain.WastePickup, int64, error) {
			return nil, 0, errors.New("db error")
		},
	}
	h := NewPickupHandler(svc)
	app.Get("/pickups", h.List)

	req := httptest.NewRequest("GET", "/pickups", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 500, resp.StatusCode)
}

func TestPickupHandler_Schedule_SafetyCheckError(t *testing.T) {
	app := fiber.New()
	svc := &mockPickupSvc{
		scheduleFn: func(id uuid.UUID, req *domain.SchedulePickupRequest) (*domain.WastePickup, error) {
			return nil, service.ErrSafetyCheckRequired
		},
	}
	h := NewPickupHandler(svc)
	app.Put("/pickups/:id/schedule", h.Schedule)

	body := `{"pickup_date":"2026-06-20T10:00:00Z"}`
	req := httptest.NewRequest("PUT", "/pickups/"+uuid.New().String()+"/schedule", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 422, resp.StatusCode)
}

func TestPickupHandler_Schedule_InvalidJSON(t *testing.T) {
	app := fiber.New()
	h := NewPickupHandler(nil)
	app.Put("/pickups/:id/schedule", h.Schedule)

	body := `{not-json}`
	req := httptest.NewRequest("PUT", "/pickups/"+uuid.New().String()+"/schedule", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 422, resp.StatusCode)
}

func TestPickupHandler_Schedule_InvalidStatus(t *testing.T) {
	app := fiber.New()
	svc := &mockPickupSvc{
		scheduleFn: func(id uuid.UUID, req *domain.SchedulePickupRequest) (*domain.WastePickup, error) {
			return nil, service.ErrInvalidStatus
		},
	}
	h := NewPickupHandler(svc)
	app.Put("/pickups/:id/schedule", h.Schedule)

	body := `{"pickup_date":"2026-06-20T10:00:00Z"}`
	req := httptest.NewRequest("PUT", "/pickups/"+uuid.New().String()+"/schedule", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 409, resp.StatusCode)
}

func TestPickupHandler_Schedule_InternalError(t *testing.T) {
	app := fiber.New()
	svc := &mockPickupSvc{
		scheduleFn: func(id uuid.UUID, req *domain.SchedulePickupRequest) (*domain.WastePickup, error) {
			return nil, errors.New("db error")
		},
	}
	h := NewPickupHandler(svc)
	app.Put("/pickups/:id/schedule", h.Schedule)

	body := `{"pickup_date":"2026-06-20T10:00:00Z"}`
	req := httptest.NewRequest("PUT", "/pickups/"+uuid.New().String()+"/schedule", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 500, resp.StatusCode)
}

func TestPickupHandler_Complete_InvalidUUID(t *testing.T) {
	app := fiber.New()
	h := NewPickupHandler(nil)
	app.Put("/pickups/:id/complete", h.Complete)

	req := httptest.NewRequest("PUT", "/pickups/not-uuid/complete", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestPickupHandler_Complete_InternalError(t *testing.T) {
	app := fiber.New()
	svc := &mockPickupSvc{
		completeFn: func(id uuid.UUID) (*domain.WastePickup, *domain.Payment, error) {
			return nil, nil, errors.New("db error")
		},
	}
	h := NewPickupHandler(svc)
	app.Put("/pickups/:id/complete", h.Complete)

	req := httptest.NewRequest("PUT", "/pickups/"+uuid.New().String()+"/complete", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 500, resp.StatusCode)
}
