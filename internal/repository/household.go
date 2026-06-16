package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
)

type HouseholdRepository interface {
	Create(h *domain.Household) error
	FindByID(id uuid.UUID) (*domain.Household, error)
	FindAll(page, perPage int) ([]domain.Household, int64, error)
	Delete(id uuid.UUID) error
}

type householdRepo struct {
	db *gorm.DB
}

func NewHouseholdRepository(db *gorm.DB) HouseholdRepository {
	return &householdRepo{db: db}
}

func (r *householdRepo) Create(h *domain.Household) error {
	return r.db.Create(h).Error
}

func (r *householdRepo) FindByID(id uuid.UUID) (*domain.Household, error) {
	var h domain.Household
	err := r.db.First(&h, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *householdRepo) FindAll(page, perPage int) ([]domain.Household, int64, error) {
	var households []domain.Household
	var total int64
	r.db.Model(&domain.Household{}).Count(&total)
	offset := (page - 1) * perPage
	err := r.db.Offset(offset).Limit(perPage).Order("created_at DESC").Find(&households).Error
	return households, total, err
}

func (r *householdRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Household{}, "id = ?", id).Error
}
