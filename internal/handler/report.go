package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
	"github.com/faisalaffan/community-waste-collection-api/internal/service"
	"github.com/faisalaffan/community-waste-collection-api/pkg/response"
)

type ReportHandler struct {
	svc service.ReportService
}

func NewReportHandler(svc service.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

// WasteSummary
// @Summary      Waste pickup summary
// @Description  Aggregated pickup counts by type and status
// @Tags         Reports
// @Produce      json
// @Success      200  {object}  response.Envelope
// @Router       /reports/waste-summary [get]
func (h *ReportHandler) WasteSummary(c fiber.Ctx) error {
	result, err := h.svc.WasteSummary()
	if err != nil {
		return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
	}

	type wasteSummaryItem struct {
		Type      string `json:"type"`
		Pending   int64  `json:"pending"`
		Completed int64  `json:"completed"`
		Canceled  int64  `json:"canceled"`
		Total     int64  `json:"total_count"`
	}
	byType := map[string]*wasteSummaryItem{}
	for _, s := range result {
		if _, ok := byType[s.Type]; !ok {
			byType[s.Type] = &wasteSummaryItem{Type: s.Type}
		}
		item := byType[s.Type]
		switch s.Status {
		case domain.PickupStatusPending:
			item.Pending = s.Count
		case domain.PickupStatusCompleted:
			item.Completed = s.Count
		case domain.PickupStatusCanceled:
			item.Canceled = s.Count
		}
		item.Total += s.Count
	}
	items := make([]wasteSummaryItem, 0, len(byType))
	for _, v := range byType {
		items = append(items, *v)
	}

	return response.SuccessOK(c, items)
}

// PaymentSummary
// @Summary      Payment summary
// @Description  Payment totals by status and total revenue
// @Tags         Reports
// @Produce      json
// @Success      200  {object}  response.Envelope
// @Router       /reports/payment-summary [get]
func (h *ReportHandler) PaymentSummary(c fiber.Ctx) error {
	result, totalRevenue, err := h.svc.PaymentSummary()
	if err != nil {
		return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
	}

	var totalPending, totalPaid, totalFailed float64
	for _, s := range result {
		switch s.Status {
		case domain.PaymentStatusPending:
			totalPending = s.TotalAmount
		case domain.PaymentStatusPaid:
			totalPaid = s.TotalAmount
		case domain.PaymentStatusFailed:
			totalFailed = s.TotalAmount
		}
	}

	return response.SuccessOK(c, fiber.Map{
		"total_pending": totalPending,
		"total_paid":    totalPaid,
		"total_failed":  totalFailed,
		"total_revenue": totalRevenue,
	})
}

// HouseholdHistory
// @Summary      Household history
// @Description  Full pickup and payment history for a household
// @Tags         Reports
// @Produce      json
// @Param        id   path      string  true  "Household UUID"
// @Success      200  {object}  response.Envelope
// @Failure      400  {object}  response.Envelope
// @Failure      404  {object}  response.Envelope
// @Router       /reports/households/{id}/history [get]
func (h *ReportHandler) HouseholdHistory(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "INVALID_ID", "id harus UUID valid")
	}

	history, err := h.svc.HouseholdHistory(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.Error(c, 404, "NOT_FOUND", "household tidak ditemukan")
		}
		return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return response.SuccessOK(c, history)
}
