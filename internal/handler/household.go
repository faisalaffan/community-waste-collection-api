package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
	"github.com/faisalaffan/community-waste-collection-api/internal/service"
	"github.com/faisalaffan/community-waste-collection-api/pkg/response"
)

type HouseholdHandler struct {
	svc service.HouseholdService
}

func NewHouseholdHandler(svc service.HouseholdService) *HouseholdHandler {
	return &HouseholdHandler{svc: svc}
}

func (h *HouseholdHandler) Create(c fiber.Ctx) error {
	var req domain.CreateHouseholdRequest
	if err := c.Bind().JSON(&req); err != nil {
		return response.ValidationError(c, "invalid request body", []response.ValidationDetail{
			{Field: "body", Message: err.Error()},
		})
	}
	if req.OwnerName == "" {
		return response.ValidationError(c, "owner_name wajib diisi", []response.ValidationDetail{
			{Field: "owner_name", Message: "wajib diisi"},
		})
	}
	if req.Address == "" {
		return response.ValidationError(c, "address wajib diisi", []response.ValidationDetail{
			{Field: "address", Message: "wajib diisi"},
		})
	}

	household, err := h.svc.Create(&req)
	if err != nil {
		return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return response.SuccessCreated(c, household)
}

func (h *HouseholdHandler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "INVALID_ID", "id harus UUID valid")
	}

	household, err := h.svc.GetByID(id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return response.Error(c, 404, "NOT_FOUND", "household tidak ditemukan")
		}
		return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return response.SuccessOK(c, household)
}

func (h *HouseholdHandler) List(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "10"))

	households, total, err := h.svc.List(page, perPage)
	if err != nil {
		return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return response.SuccessPaginated(c, households, page, perPage, total)
}

func (h *HouseholdHandler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "INVALID_ID", "id harus UUID valid")
	}

	if err := h.svc.Delete(id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return response.Error(c, 404, "NOT_FOUND", "household tidak ditemukan")
		}
		return response.Error(c, 500, "INTERNAL_ERROR", err.Error())
	}
	return response.SuccessOK(c, fiber.Map{"message": "household berhasil dihapus"})
}
