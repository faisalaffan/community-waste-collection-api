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

// Create — annotate with:
// @Summary      Create pickup request
// @Description  Create a waste pickup request (BR-01: blocks if pending payment exists)
// @Tags         Pickups
// @Accept       json
// @Produce      json
// @Param        body  body      domain.CreatePickupRequest  true  "Pickup data"
// @Success      201   {object}  response.Envelope{data=domain.WastePickup}
// @Failure      409   {object}  response.Envelope  "Pending payment exists"
// @Failure      422   {object}  response.Envelope
// @Failure      429   {object}  response.Envelope  "Rate limit exceeded"
// @Router       /pickups [post]
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

// List
// @Summary      List pickups
// @Description  List pickups with optional filters
// @Tags         Pickups
// @Produce      json
// @Param        status        query     string  false  "Filter by status"
// @Param        household_id  query     string  false  "Filter by household"
// @Param        page          query     int     false  "Page number"
// @Param        per_page      query     int     false  "Items per page"
// @Success      200           {object}  response.Envelope
// @Router       /pickups [get]
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

// Update
func (h *PickupHandler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "INVALID_ID", "id harus UUID valid")
	}

	var req domain.CreatePickupRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.ValidationError(c, "invalid request body", []response.ValidationDetail{
			{Field: "body", Message: err.Error()},
		})
	}

	pickup, err := h.svc.Update(id, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			return response.Error(c, 404, "NOT_FOUND", "pickup tidak ditemukan")
		case errors.Is(err, service.ErrPickupNotPending):
			return response.Error(c, 409, "INVALID_STATUS", err.Error())
		case errors.Is(err, service.ErrInvalidPickupType):
			return response.ValidationError(c, err.Error(), []response.ValidationDetail{
				{Field: "type", Message: "harus salah satu: organic, plastic, paper, electronic"},
			})
		}
		return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return response.SuccessOK(c, pickup)
}

// Delete
func (h *PickupHandler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "INVALID_ID", "id harus UUID valid")
	}

	if err := h.svc.Delete(id); err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			return response.Error(c, 404, "NOT_FOUND", "pickup tidak ditemukan")
		case errors.Is(err, service.ErrPickupNotPending):
			return response.Error(c, 409, "INVALID_STATUS", err.Error())
		}
		return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return response.SuccessOK(c, nil)
}

// Schedule
// @Summary      Schedule pickup
// @Description  Schedule a pending pickup (BR-02, BR-03: safety check for electronic)
// @Tags         Pickups
// @Accept       json
// @Produce      json
// @Param        id    path      string                        true  "Pickup UUID"
// @Param        body  body      domain.SchedulePickupRequest  true  "Schedule data"
// @Success      200   {object}  response.Envelope{data=domain.WastePickup}
// @Failure      400   {object}  response.Envelope
// @Failure      404   {object}  response.Envelope
// @Failure      409   {object}  response.Envelope
// @Failure      422   {object}  response.Envelope
// @Router       /pickups/{id}/schedule [put]
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

// Complete
// @Summary      Complete pickup
// @Description  Mark pickup as completed (BR-05: auto-generates payment)
// @Tags         Pickups
// @Produce      json
// @Param        id   path      string  true  "Pickup UUID"
// @Success      200  {object}  response.Envelope
// @Failure      400  {object}  response.Envelope
// @Failure      404  {object}  response.Envelope
// @Failure      409  {object}  response.Envelope
// @Router       /pickups/{id}/complete [put]
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

// Cancel
// @Summary      Cancel pickup
// @Description  Cancel a pending pickup
// @Tags         Pickups
// @Produce      json
// @Param        id   path      string  true  "Pickup UUID"
// @Success      200  {object}  response.Envelope{data=domain.WastePickup}
// @Failure      400  {object}  response.Envelope
// @Failure      404  {object}  response.Envelope
// @Failure      409  {object}  response.Envelope
// @Router       /pickups/{id}/cancel [put]
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
