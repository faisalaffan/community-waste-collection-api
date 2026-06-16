package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
)

type PaymentRepository interface {
	Create(p *domain.Payment) error
	FindByID(id uuid.UUID) (*domain.Payment, error)
	FindAll(filter PaymentFilter) ([]domain.Payment, int64, error)
	Update(p *domain.Payment) error
	HasPendingByHousehold(householdID uuid.UUID) (bool, error)
}

type PaymentFilter struct {
	HouseholdID uuid.UUID
	Status      string
	DateFrom    *time.Time
	DateTo      *time.Time
	Page        int
	PerPage     int
}

type paymentRepo struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) Create(p *domain.Payment) error {
	return r.db.Create(p).Error
}

func (r *paymentRepo) FindByID(id uuid.UUID) (*domain.Payment, error) {
	var p domain.Payment
	err := r.db.First(&p, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepo) FindAll(filter PaymentFilter) ([]domain.Payment, int64, error) {
	var payments []domain.Payment
	var total int64
	query := r.db.Model(&domain.Payment{})
	if filter.HouseholdID != uuid.Nil {
		query = query.Where("household_id = ?", filter.HouseholdID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", filter.DateFrom)
	}
	if filter.DateTo != nil {
		query = query.Where("created_at <= ?", filter.DateTo)
	}
	query.Count(&total)
	offset := (filter.Page - 1) * filter.PerPage
	err := query.Offset(offset).Limit(filter.PerPage).Order("created_at DESC").Find(&payments).Error
	return payments, total, err
}

func (r *paymentRepo) Update(p *domain.Payment) error {
	return r.db.Save(p).Error
}

func (r *paymentRepo) HasPendingByHousehold(householdID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Payment{}).
		Where("household_id = ? AND status = ?", householdID, domain.PaymentStatusPending).
		Count(&count).Error
	return count > 0, err
}
