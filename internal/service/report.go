package service

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
)

type WasteSummary struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type PaymentSummary struct {
	Status      string  `json:"status"`
	Count       int64   `json:"count"`
	TotalAmount float64 `json:"total_amount"`
}

type HouseholdHistory struct {
	Household domain.Household    `json:"household"`
	Pickups   []domain.WastePickup `json:"pickups"`
	Payments  []domain.Payment     `json:"payments"`
}

type ReportService interface {
	WasteSummary() ([]WasteSummary, error)
	PaymentSummary() ([]PaymentSummary, float64, error)
	HouseholdHistory(householdID uuid.UUID) (*HouseholdHistory, error)
	AllHistory() (*HouseholdHistory, error)
}

type reportService struct {
	db *gorm.DB
}

func NewReportService(db *gorm.DB) ReportService {
	return &reportService{db: db}
}

func (s *reportService) WasteSummary() ([]WasteSummary, error) {
	var results []WasteSummary
	err := s.db.Model(&domain.WastePickup{}).
		Select("type, status, count(*) as count").
		Group("type, status").Order("type, status").Scan(&results).Error
	return results, err
}

func (s *reportService) PaymentSummary() ([]PaymentSummary, float64, error) {
	var results []PaymentSummary
	err := s.db.Model(&domain.Payment{}).
		Select("status, count(*) as count, COALESCE(sum(amount), 0) as total_amount").
		Group("status").Order("status").Scan(&results).Error
	if err != nil {
		return nil, 0, err
	}
	var totalRevenue float64
	s.db.Model(&domain.Payment{}).Where("status = ?", domain.PaymentStatusPaid).
		Select("COALESCE(sum(amount), 0)").Scan(&totalRevenue)
	return results, totalRevenue, nil
}

func (s *reportService) HouseholdHistory(householdID uuid.UUID) (*HouseholdHistory, error) {
	var h domain.Household
	if err := s.db.First(&h, "id = ?", householdID).Error; err != nil {
		return nil, err
	}
	var pickups []domain.WastePickup
	s.db.Where("household_id = ?", householdID).Order("created_at DESC").Find(&pickups)
	var payments []domain.Payment
	s.db.Where("household_id = ?", householdID).Order("created_at DESC").Find(&payments)

	return &HouseholdHistory{
		Household: h,
		Pickups:   pickups,
		Payments:  payments,
	}, nil
}

func (s *reportService) AllHistory() (*HouseholdHistory, error) {
	var pickups []domain.WastePickup
	if err := s.db.Order("created_at DESC").Find(&pickups).Error; err != nil {
		return nil, err
	}
	var payments []domain.Payment
	if err := s.db.Order("created_at DESC").Find(&payments).Error; err != nil {
		return nil, err
	}

	return &HouseholdHistory{
		Household: domain.Household{
			OwnerName: "Semua Warga",
			Address:   "Komunitas WasteCo",
		},
		Pickups:   pickups,
		Payments:  payments,
	}, nil
}
