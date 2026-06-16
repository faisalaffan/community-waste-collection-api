package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
	"github.com/faisalaffan/community-waste-collection-api/internal/repository"
	"github.com/faisalaffan/community-waste-collection-api/internal/service"
	"github.com/faisalaffan/community-waste-collection-api/pkg/response"
)

type PickupHandler struct {
	svc service.PickupService
}

func NewPickupHandler(svc service.PickupService) *PickupHandler {
	return &PickupHandler{svc: svc}
}

func (h *PickupHandler) Create(c fiber.Ctx) error {
	var req domain.CreatePickupRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.ValidationError(c, "invalid request body", []response.ValidationDetail{
			{Field: "body", Message: err.Error()},
		})
	}

	validTypes := map[string]bool{"organic": true, "plastic": true, "paper": true, "electronic": true}
	if !validTypes[req.Type] {
		return response.ValidationError(c, "type tidak valid", []response.ValidationDetail{
			{Field: "type", Message: "harus salah satu: organic, plastic, paper, electronic"},
		})
	}
	if req.HouseholdID == uuid.Nil {
		return response.ValidationError(c, "household_id wajib diisi", []response.ValidationDetail{
			{Field: "household_id", Message: "wajib diisi"},
		})
	}

	pickup, err := h.svc.Create(&req)
	if err != nil {
		if errors.Is(err, service.ErrPendingPayment) {
			return response.Error(c, 409, "PICKUP_BLOCKED", err.Error())
		}
		return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return response.SuccessCreated(c, pickup)
}

func (h *PickupHandler) List(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "10"))
	householdID, _ := uuid.Parse(c.Query("household_id", ""))
	status := c.Query("status", "")

	pickups, total, err := h.svc.List(repository.PickupFilter{
		HouseholdID: householdID,
		Status:      status,
		Page:        page,
		PerPage:     perPage,
	})
	if err != nil {
		return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return response.SuccessPaginated(c, pickups, page, perPage, total)
}

func (h *PickupHandler) Schedule(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "INVALID_ID", "id harus UUID valid")
	}

	var req domain.SchedulePickupRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.ValidationError(c, "invalid request body", []response.ValidationDetail{
			{Field: "body", Message: err.Error()},
		})
	}

	pickup, err := h.svc.Schedule(id, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			return response.Error(c, 404, "NOT_FOUND", "pickup tidak ditemukan")
		case errors.Is(err, service.ErrInvalidStatus):
			return response.Error(c, 409, "INVALID_STATUS", err.Error())
		case errors.Is(err, service.ErrSafetyCheckRequired):
			return response.ValidationError(c, err.Error(), []response.ValidationDetail{
				{Field: "safety_check", Message: "safety_check harus true untuk electronic waste"},
			})
		}
		return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return response.SuccessOK(c, pickup)
}

func (h *PickupHandler) Complete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "INVALID_ID", "id harus UUID valid")
	}

	pickup, payment, err := h.svc.Complete(id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return response.Error(c, 404, "NOT_FOUND", "pickup tidak ditemukan")
		}
		return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return response.SuccessOK(c, fiber.Map{
		"pickup":  pickup,
		"payment": payment,
	})
}

func (h *PickupHandler) Cancel(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "INVALID_ID", "id harus UUID valid")
	}

	pickup, err := h.svc.Cancel(id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return response.Error(c, 404, "NOT_FOUND", "pickup tidak ditemukan")
		}
		if errors.Is(err, service.ErrPickupNotPending) {
			return response.Error(c, 409, "INVALID_STATUS", err.Error())
		}
		return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return response.SuccessOK(c, pickup)
}
