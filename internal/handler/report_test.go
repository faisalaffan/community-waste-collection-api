package handler

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/faisalaffan/community-waste-collection-api/internal/service"
)

type mockReportSvc struct {
	wasteSummaryFn    func() ([]service.WasteSummary, error)
	paymentSummaryFn  func() ([]service.PaymentSummary, float64, error)
	householdHistoryFn func(id uuid.UUID) (*service.HouseholdHistory, error)
	allHistoryFn      func() (*service.HouseholdHistory, error)
}

func (m *mockReportSvc) WasteSummary() ([]service.WasteSummary, error) {
	return m.wasteSummaryFn()
}
func (m *mockReportSvc) PaymentSummary() ([]service.PaymentSummary, float64, error) {
	return m.paymentSummaryFn()
}
func (m *mockReportSvc) HouseholdHistory(id uuid.UUID) (*service.HouseholdHistory, error) {
	return m.householdHistoryFn(id)
}
func (m *mockReportSvc) AllHistory() (*service.HouseholdHistory, error) {
	if m.allHistoryFn != nil {
		return m.allHistoryFn()
	}
	return &service.HouseholdHistory{}, nil
}

func TestReportHandler_WasteSummary_Success(t *testing.T) {
	app := fiber.New()
	svc := &mockReportSvc{
		wasteSummaryFn: func() ([]service.WasteSummary, error) {
			return []service.WasteSummary{
				{Type: "organic", Status: "pending", Count: 5},
			}, nil
		},
	}
	h := NewReportHandler(svc)
	app.Get("/reports/waste-summary", h.WasteSummary)

	req := httptest.NewRequest("GET", "/reports/waste-summary", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestReportHandler_WasteSummary_InternalError(t *testing.T) {
	app := fiber.New()
	svc := &mockReportSvc{
		wasteSummaryFn: func() ([]service.WasteSummary, error) {
			return nil, errors.New("db error")
		},
	}
	h := NewReportHandler(svc)
	app.Get("/reports/waste-summary", h.WasteSummary)

	req := httptest.NewRequest("GET", "/reports/waste-summary", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 500, resp.StatusCode)
}

func TestReportHandler_PaymentSummary_Success(t *testing.T) {
	app := fiber.New()
	svc := &mockReportSvc{
		paymentSummaryFn: func() ([]service.PaymentSummary, float64, error) {
			return []service.PaymentSummary{
				{Status: "paid", Count: 1, TotalAmount: 50000},
			}, 50000.0, nil
		},
	}
	h := NewReportHandler(svc)
	app.Get("/reports/payment-summary", h.PaymentSummary)

	req := httptest.NewRequest("GET", "/reports/payment-summary", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestReportHandler_PaymentSummary_InternalError(t *testing.T) {
	app := fiber.New()
	svc := &mockReportSvc{
		paymentSummaryFn: func() ([]service.PaymentSummary, float64, error) {
			return nil, 0, errors.New("db error")
		},
	}
	h := NewReportHandler(svc)
	app.Get("/reports/payment-summary", h.PaymentSummary)

	req := httptest.NewRequest("GET", "/reports/payment-summary", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 500, resp.StatusCode)
}

func TestReportHandler_HouseholdHistory_Success(t *testing.T) {
	app := fiber.New()
	svc := &mockReportSvc{
		householdHistoryFn: func(id uuid.UUID) (*service.HouseholdHistory, error) {
			return &service.HouseholdHistory{}, nil
		},
	}
	h := NewReportHandler(svc)
	app.Get("/reports/households/:id/history", h.HouseholdHistory)

	req := httptest.NewRequest("GET", "/reports/households/"+uuid.New().String()+"/history", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestReportHandler_HouseholdHistory_InvalidUUID(t *testing.T) {
	app := fiber.New()
	h := NewReportHandler(nil)
	app.Get("/reports/households/:id/history", h.HouseholdHistory)

	req := httptest.NewRequest("GET", "/reports/households/not-uuid/history", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestReportHandler_HouseholdHistory_NotFound(t *testing.T) {
	app := fiber.New()
	svc := &mockReportSvc{
		householdHistoryFn: func(id uuid.UUID) (*service.HouseholdHistory, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	h := NewReportHandler(svc)
	app.Get("/reports/households/:id/history", h.HouseholdHistory)

	req := httptest.NewRequest("GET", "/reports/households/"+uuid.New().String()+"/history", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestReportHandler_HouseholdHistory_InternalError(t *testing.T) {
	app := fiber.New()
	svc := &mockReportSvc{
		householdHistoryFn: func(id uuid.UUID) (*service.HouseholdHistory, error) {
			return nil, errors.New("db error")
		},
	}
	h := NewReportHandler(svc)
	app.Get("/reports/households/:id/history", h.HouseholdHistory)

	req := httptest.NewRequest("GET", "/reports/households/"+uuid.New().String()+"/history", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 500, resp.StatusCode)
}
