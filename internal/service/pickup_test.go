package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
	"github.com/faisalaffan/community-waste-collection-api/internal/repository"
)

type mockPickupRepo struct {
	createFn               func(p *domain.WastePickup) error
	findByIDFn             func(id uuid.UUID) (*domain.WastePickup, error)
	findAllFn              func(filter repository.PickupFilter) ([]domain.WastePickup, int64, error)
	updateFn               func(p *domain.WastePickup) error
	cancelOrganicPendingFn func(olderThan time.Duration) (int64, error)
}

func (m *mockPickupRepo) Create(p *domain.WastePickup) error { return m.createFn(p) }
func (m *mockPickupRepo) FindByID(id uuid.UUID) (*domain.WastePickup, error) {
	return m.findByIDFn(id)
}
func (m *mockPickupRepo) FindAll(filter repository.PickupFilter) ([]domain.WastePickup, int64, error) {
	return m.findAllFn(filter)
}
func (m *mockPickupRepo) Update(p *domain.WastePickup) error { return m.updateFn(p) }
func (m *mockPickupRepo) CancelOrganicPending(olderThan time.Duration) (int64, error) {
	return m.cancelOrganicPendingFn(olderThan)
}

type mockPaymentRepo struct {
	createFn                func(p *domain.Payment) error
	findByIDFn              func(id uuid.UUID) (*domain.Payment, error)
	findAllFn               func(filter repository.PaymentFilter) ([]domain.Payment, int64, error)
	updateFn                func(p *domain.Payment) error
	hasPendingByHouseholdFn func(householdID uuid.UUID) (bool, error)
}

func (m *mockPaymentRepo) Create(p *domain.Payment) error { return m.createFn(p) }
func (m *mockPaymentRepo) FindByID(id uuid.UUID) (*domain.Payment, error) {
	return m.findByIDFn(id)
}
func (m *mockPaymentRepo) FindAll(filter repository.PaymentFilter) ([]domain.Payment, int64, error) {
	return m.findAllFn(filter)
}
func (m *mockPaymentRepo) Update(p *domain.Payment) error { return m.updateFn(p) }
func (m *mockPaymentRepo) HasPendingByHousehold(householdID uuid.UUID) (bool, error) {
	return m.hasPendingByHouseholdFn(householdID)
}

func TestPickupService_Create_BR01_Blocked(t *testing.T) {
	pmr := &mockPaymentRepo{
		hasPendingByHouseholdFn: func(householdID uuid.UUID) (bool, error) { return true, nil },
	}
	pr := &mockPickupRepo{}
	svc := NewPickupService(pr, pmr)
	_, err := svc.Create(&domain.CreatePickupRequest{HouseholdID: uuid.New(), Type: domain.PickupTypeOrganic})
	assert.Equal(t, ErrPendingPayment, err)
}

func TestPickupService_Create_Success(t *testing.T) {
	pmr := &mockPaymentRepo{
		hasPendingByHouseholdFn: func(householdID uuid.UUID) (bool, error) { return false, nil },
	}
	pr := &mockPickupRepo{
		createFn: func(p *domain.WastePickup) error { return nil },
	}
	svc := NewPickupService(pr, pmr)
	p, err := svc.Create(&domain.CreatePickupRequest{HouseholdID: uuid.New(), Type: domain.PickupTypeOrganic})
	assert.NoError(t, err)
	assert.Equal(t, domain.PickupStatusPending, p.Status)
}

func TestPickupService_Schedule_BR02_InvalidStatus(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{ID: id, Status: domain.PickupStatusCompleted}, nil
		},
	}
	svc := NewPickupService(pr, nil)
	_, err := svc.Schedule(uuid.New(), &domain.SchedulePickupRequest{PickupDate: time.Now()})
	assert.Equal(t, ErrInvalidStatus, err)
}

func TestPickupService_Schedule_BR03_SafetyCheck(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{ID: id, Type: domain.PickupTypeElectronic, Status: domain.PickupStatusPending, SafetyCheck: false}, nil
		},
	}
	svc := NewPickupService(pr, nil)
	_, err := svc.Schedule(uuid.New(), &domain.SchedulePickupRequest{PickupDate: time.Now()})
	assert.Equal(t, ErrSafetyCheckRequired, err)
}

func TestPickupService_Complete_BR05_GeneratesPayment(t *testing.T) {
	pickupID := uuid.New()
	householdID := uuid.New()
	pickup := &domain.WastePickup{
		ID: pickupID, HouseholdID: householdID, Type: domain.PickupTypeElectronic, Status: domain.PickupStatusScheduled,
	}
	pr := &mockPickupRepo{
		findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) { return pickup, nil },
		updateFn:   func(p *domain.WastePickup) error { return nil },
	}
	pmr := &mockPaymentRepo{
		createFn: func(p *domain.Payment) error {
			assert.Equal(t, 100000.0, p.Amount)
			return nil
		},
	}
	svc := NewPickupService(pr, pmr)
	_, payment, err := svc.Complete(pickupID)
	assert.NoError(t, err)
	assert.Equal(t, householdID, payment.HouseholdID)
	assert.Equal(t, pickupID, payment.WasteID)
}
