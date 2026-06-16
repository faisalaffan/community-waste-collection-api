package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
	"github.com/faisalaffan/community-waste-collection-api/internal/repository"
)

var (
	ErrPendingPayment       = errors.New("household masih memiliki pending payment")
	ErrInvalidStatus        = errors.New("pickup hanya dapat dijadwalkan saat status pending")
	ErrSafetyCheckRequired  = errors.New("pickup electronic memerlukan safety_check = true")
	ErrPickupNotPending     = errors.New("hanya pickup dengan status pending yang dapat dibatalkan")
	ErrPickupNotSchedulable = errors.New("pickup hanya dapat diselesaikan saat status scheduled")
	ErrInvalidPickupType    = errors.New("tipe pickup tidak valid")
)

type PickupService interface {
	Create(req *domain.CreatePickupRequest) (*domain.WastePickup, error)
	GetByID(id uuid.UUID) (*domain.WastePickup, error)
	List(filter repository.PickupFilter) ([]domain.WastePickup, int64, error)
	Schedule(id uuid.UUID, req *domain.SchedulePickupRequest) (*domain.WastePickup, error)
	Complete(id uuid.UUID) (*domain.WastePickup, *domain.Payment, error)
	Cancel(id uuid.UUID) (*domain.WastePickup, error)
}

type pickupService struct {
	pickupRepo  repository.PickupRepository
	paymentRepo repository.PaymentRepository
}

func NewPickupService(pr repository.PickupRepository, pmr repository.PaymentRepository) PickupService {
	return &pickupService{pickupRepo: pr, paymentRepo: pmr}
}

// BR-01: Blokir pickup jika household punya pending payment
func (s *pickupService) Create(req *domain.CreatePickupRequest) (*domain.WastePickup, error) {
	hasPending, err := s.paymentRepo.HasPendingByHousehold(req.HouseholdID)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, ErrPendingPayment
	}

	safetyCheck := false
	if req.SafetyCheck != nil {
		safetyCheck = *req.SafetyCheck
	}

	p := &domain.WastePickup{
		ID:          uuid.New(),
		HouseholdID: req.HouseholdID,
		Type:        req.Type,
		Status:      domain.PickupStatusPending,
		SafetyCheck: safetyCheck,
	}
	if err := s.pickupRepo.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *pickupService) GetByID(id uuid.UUID) (*domain.WastePickup, error) {
	p, err := s.pickupRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *pickupService) List(filter repository.PickupFilter) ([]domain.WastePickup, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 10
	}
	return s.pickupRepo.FindAll(filter)
}

// BR-02: Only pending -> scheduled
// BR-03: Safety check untuk electronic
func (s *pickupService) Schedule(id uuid.UUID, req *domain.SchedulePickupRequest) (*domain.WastePickup, error) {
	p, err := s.pickupRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if p.Status != domain.PickupStatusPending {
		return nil, ErrInvalidStatus
	}

	if p.Type == domain.PickupTypeElectronic && !p.SafetyCheck {
		return nil, ErrSafetyCheckRequired
	}

	p.Status = domain.PickupStatusScheduled
	p.PickupDate = &req.PickupDate
	if err := s.pickupRepo.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

// BR-05: Auto-generate payment saat completed
func (s *pickupService) Complete(id uuid.UUID) (*domain.WastePickup, *domain.Payment, error) {
	p, err := s.pickupRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}

	if p.Status != domain.PickupStatusScheduled {
		return nil, nil, ErrPickupNotSchedulable
	}

	amount, ok := domain.PickupAmounts[p.Type]
	if !ok {
		return nil, nil, ErrInvalidPickupType
	}

	p.Status = domain.PickupStatusCompleted
	if p.PickupDate == nil {
		now := time.Now()
		p.PickupDate = &now
	}
	if err := s.pickupRepo.Update(p); err != nil {
		return nil, nil, err
	}
	payment := &domain.Payment{
		ID:          uuid.New(),
		HouseholdID: p.HouseholdID,
		WasteID:     p.ID,
		Amount:      amount,
		Status:      domain.PaymentStatusPending,
	}
	if err := s.paymentRepo.Create(payment); err != nil {
		return nil, nil, err
	}

	return p, payment, nil
}

func (s *pickupService) Cancel(id uuid.UUID) (*domain.WastePickup, error) {
	p, err := s.pickupRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if p.Status != domain.PickupStatusPending {
		return nil, ErrPickupNotPending
	}
	p.Status = domain.PickupStatusCanceled
	if err := s.pickupRepo.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}
