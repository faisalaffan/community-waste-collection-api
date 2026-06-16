package handler

import (
	"bytes"
	"errors"
	"mime/multipart"
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

type mockPaymentSvc struct {
	createFn  func(req *domain.CreatePaymentRequest) (*domain.Payment, error)
	getByIDFn func(id uuid.UUID) (*domain.Payment, error)
	listFn    func(filter repository.PaymentFilter) ([]domain.Payment, int64, error)
	confirmFn func(id uuid.UUID, file *multipart.FileHeader) (*domain.Payment, error)
}

func (m *mockPaymentSvc) Create(req *domain.CreatePaymentRequest) (*domain.Payment, error) {
	return m.createFn(req)
}
func (m *mockPaymentSvc) GetByID(id uuid.UUID) (*domain.Payment, error) {
	return m.getByIDFn(id)
}
func (m *mockPaymentSvc) List(filter repository.PaymentFilter) ([]domain.Payment, int64, error) {
	return m.listFn(filter)
}
func (m *mockPaymentSvc) Confirm(id uuid.UUID, file *multipart.FileHeader) (*domain.Payment, error) {
	return m.confirmFn(id, file)
}

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

func TestPaymentHandler_Create_InvalidJSON(t *testing.T) {
	app := fiber.New()
	h := NewPaymentHandler(nil)
	app.Post("/payments", h.Create)

	body := `{not-json}`
	req := httptest.NewRequest("POST", "/payments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 422, resp.StatusCode)
}

func TestPaymentHandler_Create_NilHouseholdID(t *testing.T) {
	app := fiber.New()
	h := NewPaymentHandler(nil)
	app.Post("/payments", h.Create)

	body := `{"household_id":"00000000-0000-0000-0000-000000000000"}`
	req := httptest.NewRequest("POST", "/payments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 422, resp.StatusCode)
}

func TestPaymentHandler_Create_Success(t *testing.T) {
	app := fiber.New()
	svc := &mockPaymentSvc{
		createFn: func(req *domain.CreatePaymentRequest) (*domain.Payment, error) {
			return &domain.Payment{
				ID:          uuid.New(),
				HouseholdID: req.HouseholdID,
				Amount:      0,
				Status:      domain.PaymentStatusPending,
			}, nil
		},
	}
	h := NewPaymentHandler(svc)
	app.Post("/payments", h.Create)

	body := `{"household_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest("POST", "/payments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 201, resp.StatusCode)
}

func TestPaymentHandler_Create_InternalError(t *testing.T) {
	app := fiber.New()
	svc := &mockPaymentSvc{
		createFn: func(req *domain.CreatePaymentRequest) (*domain.Payment, error) {
			return nil, errors.New("db error")
		},
	}
	h := NewPaymentHandler(svc)
	app.Post("/payments", h.Create)

	body := `{"household_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest("POST", "/payments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 500, resp.StatusCode)
}

func TestPaymentHandler_List_Success(t *testing.T) {
	app := fiber.New()
	svc := &mockPaymentSvc{
		listFn: func(filter repository.PaymentFilter) ([]domain.Payment, int64, error) {
			return []domain.Payment{}, 0, nil
		},
	}
	h := NewPaymentHandler(svc)
	app.Get("/payments", h.List)

	req := httptest.NewRequest("GET", "/payments", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestPaymentHandler_List_WithDateFilter(t *testing.T) {
	app := fiber.New()
	svc := &mockPaymentSvc{
		listFn: func(filter repository.PaymentFilter) ([]domain.Payment, int64, error) {
			assert.NotNil(t, filter.DateFrom)
			assert.NotNil(t, filter.DateTo)
			return []domain.Payment{}, 0, nil
		},
	}
	h := NewPaymentHandler(svc)
	app.Get("/payments", h.List)

	req := httptest.NewRequest("GET", "/payments?date_from=2026-01-01&date_to=2026-06-01", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestPaymentHandler_List_InvalidDateFrom(t *testing.T) {
	app := fiber.New()
	h := NewPaymentHandler(nil)
	app.Get("/payments", h.List)

	req := httptest.NewRequest("GET", "/payments?date_from=not-a-date", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestPaymentHandler_List_InvalidDateTo(t *testing.T) {
	app := fiber.New()
	h := NewPaymentHandler(nil)
	app.Get("/payments", h.List)

	req := httptest.NewRequest("GET", "/payments?date_to=not-a-date", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestPaymentHandler_List_InternalError(t *testing.T) {
	app := fiber.New()
	svc := &mockPaymentSvc{
		listFn: func(filter repository.PaymentFilter) ([]domain.Payment, int64, error) {
			return nil, 0, errors.New("db error")
		},
	}
	h := NewPaymentHandler(svc)
	app.Get("/payments", h.List)

	req := httptest.NewRequest("GET", "/payments", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 500, resp.StatusCode)
}

func TestPaymentHandler_Confirm_InvalidUUID(t *testing.T) {
	app := fiber.New()
	h := NewPaymentHandler(nil)
	app.Put("/payments/:id/confirm", h.Confirm)

	req := httptest.NewRequest("PUT", "/payments/not-uuid/confirm", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestPaymentHandler_Confirm_MissingFile(t *testing.T) {
	app := fiber.New()
	svc := &mockPaymentSvc{
		confirmFn: func(id uuid.UUID, file *multipart.FileHeader) (*domain.Payment, error) {
			return nil, errors.New("file bukti pembayaran wajib diupload")
		},
	}
	h := NewPaymentHandler(svc)
	app.Put("/payments/:id/confirm", h.Confirm)

	req := httptest.NewRequest("PUT", "/payments/"+uuid.New().String()+"/confirm", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 422, resp.StatusCode)
}

func TestPaymentHandler_Confirm_NotFound(t *testing.T) {
	app := fiber.New()
	svc := &mockPaymentSvc{
		confirmFn: func(id uuid.UUID, file *multipart.FileHeader) (*domain.Payment, error) {
			return nil, service.ErrPaymentNotFound
		},
	}
	h := NewPaymentHandler(svc)
	app.Put("/payments/:id/confirm", h.Confirm)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("proof_file", "proof.jpg")
	part.Write([]byte("fake-image-data"))
	writer.Close()

	req := httptest.NewRequest("PUT", "/payments/"+uuid.New().String()+"/confirm", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, _ := app.Test(req)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestPaymentHandler_Confirm_NotPending(t *testing.T) {
	app := fiber.New()
	svc := &mockPaymentSvc{
		confirmFn: func(id uuid.UUID, file *multipart.FileHeader) (*domain.Payment, error) {
			return nil, service.ErrPaymentNotPending
		},
	}
	h := NewPaymentHandler(svc)
	app.Put("/payments/:id/confirm", h.Confirm)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("proof_file", "proof.jpg")
	part.Write([]byte("fake-image-data"))
	writer.Close()

	req := httptest.NewRequest("PUT", "/payments/"+uuid.New().String()+"/confirm", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, _ := app.Test(req)
	assert.Equal(t, 409, resp.StatusCode)
}

func TestPaymentHandler_Confirm_ProofFileRequired(t *testing.T) {
	app := fiber.New()
	svc := &mockPaymentSvc{
		confirmFn: func(id uuid.UUID, file *multipart.FileHeader) (*domain.Payment, error) {
			return nil, service.ErrProofFileRequired
		},
	}
	h := NewPaymentHandler(svc)
	app.Put("/payments/:id/confirm", h.Confirm)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("proof_file", "proof.jpg")
	part.Write([]byte("fake-image-data"))
	writer.Close()

	req := httptest.NewRequest("PUT", "/payments/"+uuid.New().String()+"/confirm", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, _ := app.Test(req)
	assert.Equal(t, 422, resp.StatusCode)
}

func TestPaymentHandler_Confirm_InternalError(t *testing.T) {
	app := fiber.New()
	svc := &mockPaymentSvc{
		confirmFn: func(id uuid.UUID, file *multipart.FileHeader) (*domain.Payment, error) {
			return nil, errors.New("db error")
		},
	}
	h := NewPaymentHandler(svc)
	app.Put("/payments/:id/confirm", h.Confirm)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("proof_file", "proof.jpg")
	part.Write([]byte("fake-image-data"))
	writer.Close()

	req := httptest.NewRequest("PUT", "/payments/"+uuid.New().String()+"/confirm", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, _ := app.Test(req)
	assert.Equal(t, 500, resp.StatusCode)
}

func TestPaymentHandler_Confirm_Success(t *testing.T) {
	app := fiber.New()
	svc := &mockPaymentSvc{
		confirmFn: func(id uuid.UUID, file *multipart.FileHeader) (*domain.Payment, error) {
			return &domain.Payment{ID: id, Amount: 50000, Status: domain.PaymentStatusPaid}, nil
		},
	}
	h := NewPaymentHandler(svc)
	app.Put("/payments/:id/confirm", h.Confirm)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("proof_file", "proof.jpg")
	part.Write([]byte("fake-image-data"))
	writer.Close()

	req := httptest.NewRequest("PUT", "/payments/"+uuid.New().String()+"/confirm", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}
