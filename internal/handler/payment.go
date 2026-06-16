package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
	"github.com/faisalaffan/community-waste-collection-api/internal/repository"
	"github.com/faisalaffan/community-waste-collection-api/internal/service"
	"github.com/faisalaffan/community-waste-collection-api/pkg/response"
)

type PaymentHandler struct {
	svc service.PaymentService
}

func NewPaymentHandler(svc service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

func (h *PaymentHandler) Create(c fiber.Ctx) error {
	var req domain.CreatePaymentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.ValidationError(c, "invalid request body", []response.ValidationDetail{
			{Field: "body", Message: err.Error()},
		})
	}
	if req.HouseholdID == uuid.Nil {
		return response.ValidationError(c, "household_id wajib diisi", []response.ValidationDetail{
			{Field: "household_id", Message: "wajib diisi"},
		})
	}

	payment, err := h.svc.Create(&req)
	if err != nil {
		return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return response.SuccessCreated(c, payment)
}

func (h *PaymentHandler) List(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "10"))
	householdID, _ := uuid.Parse(c.Query("household_id", ""))
	status := c.Query("status", "")

	filter := repository.PaymentFilter{
		HouseholdID: householdID,
		Status:      status,
		Page:        page,
		PerPage:     perPage,
	}

	if dateFrom := c.Query("date_from", ""); dateFrom != "" {
		t, err := time.Parse("2006-01-02", dateFrom)
		if err != nil {
			return response.Error(c, 400, "INVALID_DATE", "format date_from harus YYYY-MM-DD")
		}
		filter.DateFrom = &t
	}
	if dateTo := c.Query("date_to", ""); dateTo != "" {
		t, err := time.Parse("2006-01-02", dateTo)
		if err != nil {
			return response.Error(c, 400, "INVALID_DATE", "format date_to harus YYYY-MM-DD")
		}
		filter.DateTo = &t
	}

	payments, total, err := h.svc.List(filter)
	if err != nil {
		return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return response.SuccessPaginated(c, payments, page, perPage, total)
}

func (h *PaymentHandler) Confirm(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "INVALID_ID", "id harus UUID valid")
	}

	file, err := c.FormFile("proof_file")
	if err != nil {
		return response.ValidationError(c, "file bukti pembayaran wajib diupload", []response.ValidationDetail{
			{Field: "proof_file", Message: "wajib diupload"},
		})
	}

	payment, err := h.svc.Confirm(id, file)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPaymentNotFound):
			return response.Error(c, 404, "NOT_FOUND", "payment tidak ditemukan")
		case errors.Is(err, service.ErrPaymentNotPending):
			return response.Error(c, 409, "INVALID_STATUS", err.Error())
		case errors.Is(err, service.ErrProofFileRequired):
			return response.ValidationError(c, err.Error(), []response.ValidationDetail{
				{Field: "proof_file", Message: "wajib diupload"},
			})
		}
		return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return response.SuccessOK(c, payment)
}
