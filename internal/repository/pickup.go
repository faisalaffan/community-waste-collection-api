package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
)

type PickupRepository interface {
	Create(p *domain.WastePickup) error
	FindByID(id uuid.UUID) (*domain.WastePickup, error)
	FindAll(filter PickupFilter) ([]domain.WastePickup, int64, error)
	Update(p *domain.WastePickup) error
	CancelOrganicPending(olderThan time.Duration) (int64, error)
}

type PickupFilter struct {
	HouseholdID uuid.UUID
	Status      string
	Page        int
	PerPage     int
}

type pickupRepo struct {
	db *gorm.DB
}

func NewPickupRepository(db *gorm.DB) PickupRepository {
	return &pickupRepo{db: db}
}

func (r *pickupRepo) Create(p *domain.WastePickup) error {
	return r.db.Create(p).Error
}

func (r *pickupRepo) FindByID(id uuid.UUID) (*domain.WastePickup, error) {
	var p domain.WastePickup
	err := r.db.First(&p, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *pickupRepo) FindAll(filter PickupFilter) ([]domain.WastePickup, int64, error) {
	var pickups []domain.WastePickup
	var total int64
	query := r.db.Model(&domain.WastePickup{})
	if filter.HouseholdID != uuid.Nil {
		query = query.Where("household_id = ?", filter.HouseholdID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	query.Count(&total)
	offset := (filter.Page - 1) * filter.PerPage
	err := query.Offset(offset).Limit(filter.PerPage).Order("created_at DESC").Find(&pickups).Error
	return pickups, total, err
}

func (r *pickupRepo) Update(p *domain.WastePickup) error {
	return r.db.Save(p).Error
}

func (r *pickupRepo) CancelOrganicPending(olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	result := r.db.Model(&domain.WastePickup{}).
		Where("type = ? AND status = ? AND created_at < ?", domain.PickupTypeOrganic, domain.PickupStatusPending, cutoff).
		Update("status", domain.PickupStatusCanceled)
	return result.RowsAffected, result.Error
}
