package service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
	"github.com/faisalaffan/community-waste-collection-api/internal/repository"
)

type mockPickupRepo struct {
	createFn               func(p *domain.WastePickup) error
	findByIDFn             func(id uuid.UUID) (*domain.WastePickup, error)
	findAllFn              func(filter repository.PickupFilter) ([]domain.WastePickup, int64, error)
	updateFn               func(p *domain.WastePickup) error
	deleteFn               func(id uuid.UUID) error
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
func (m *mockPickupRepo) Delete(id uuid.UUID) error          { return m.deleteFn(id) }
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

func TestPickupService_Create_RepoError(t *testing.T) {
	pmr := &mockPaymentRepo{
		hasPendingByHouseholdFn: func(householdID uuid.UUID) (bool, error) { return false, nil },
	}
	pr := &mockPickupRepo{
		createFn: func(p *domain.WastePickup) error { return errors.New("db error") },
	}
	svc := NewPickupService(pr, pmr)
	_, err := svc.Create(&domain.CreatePickupRequest{HouseholdID: uuid.New(), Type: domain.PickupTypeOrganic})
	assert.Error(t, err)
}

func TestPickupService_Create_HasPendingError(t *testing.T) {
	pmr := &mockPaymentRepo{
		hasPendingByHouseholdFn: func(householdID uuid.UUID) (bool, error) {
			return false, errors.New("db error")
		},
	}
	pr := &mockPickupRepo{}
	svc := NewPickupService(pr, pmr)
	_, err := svc.Create(&domain.CreatePickupRequest{HouseholdID: uuid.New(), Type: domain.PickupTypeOrganic})
	assert.Error(t, err)
}

func TestPickupService_Create_WithSafetyCheck(t *testing.T) {
	safety := true
	pmr := &mockPaymentRepo{
		hasPendingByHouseholdFn: func(householdID uuid.UUID) (bool, error) { return false, nil },
	}
	pr := &mockPickupRepo{
		createFn: func(p *domain.WastePickup) error {
			assert.True(t, p.SafetyCheck)
			return nil
		},
	}
	svc := NewPickupService(pr, pmr)
	p, err := svc.Create(&domain.CreatePickupRequest{
		HouseholdID: uuid.New(), Type: domain.PickupTypeElectronic, SafetyCheck: &safety,
	})
	assert.NoError(t, err)
	assert.True(t, p.SafetyCheck)
}

func TestPickupService_GetByID_Success(t *testing.T) {
	id := uuid.New()
	pr := &mockPickupRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{ID: uid, Type: domain.PickupTypeOrganic}, nil
		},
	}
	svc := NewPickupService(pr, nil)
	p, err := svc.GetByID(id)
	assert.NoError(t, err)
	assert.Equal(t, id, p.ID)
	assert.Equal(t, domain.PickupTypeOrganic, p.Type)
}

func TestPickupService_GetByID_NotFound(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	svc := NewPickupService(pr, nil)
	_, err := svc.GetByID(uuid.New())
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestPickupService_GetByID_RepoError(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewPickupService(pr, nil)
	_, err := svc.GetByID(uuid.New())
	assert.Error(t, err)
	assert.False(t, errors.Is(err, ErrNotFound))
}

func TestPickupService_List_Defaults(t *testing.T) {
	pr := &mockPickupRepo{
		findAllFn: func(filter repository.PickupFilter) ([]domain.WastePickup, int64, error) {
			assert.Equal(t, 1, filter.Page)
			assert.Equal(t, 10, filter.PerPage)
			return []domain.WastePickup{}, 0, nil
		},
	}
	svc := NewPickupService(pr, nil)
	svc.List(repository.PickupFilter{Page: 0, PerPage: 0})
}

func TestPickupService_List_FilterPassthrough(t *testing.T) {
	hID := uuid.New()
	pr := &mockPickupRepo{
		findAllFn: func(filter repository.PickupFilter) ([]domain.WastePickup, int64, error) {
			assert.Equal(t, hID, filter.HouseholdID)
			assert.Equal(t, "pending", filter.Status)
			return []domain.WastePickup{}, 0, nil
		},
	}
	svc := NewPickupService(pr, nil)
	svc.List(repository.PickupFilter{HouseholdID: hID, Status: "pending", Page: 1, PerPage: 10})
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

func TestPickupService_Schedule_Success(t *testing.T) {
	now := time.Now()
	pr := &mockPickupRepo{
		findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{ID: id, Type: domain.PickupTypeOrganic, Status: domain.PickupStatusPending}, nil
		},
		updateFn: func(p *domain.WastePickup) error { return nil },
	}
	svc := NewPickupService(pr, nil)
	p, err := svc.Schedule(uuid.New(), &domain.SchedulePickupRequest{PickupDate: now})
	assert.NoError(t, err)
	assert.Equal(t, domain.PickupStatusScheduled, p.Status)
	assert.Equal(t, &now, p.PickupDate)
}

func TestPickupService_Schedule_NotFound(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	svc := NewPickupService(pr, nil)
	_, err := svc.Schedule(uuid.New(), &domain.SchedulePickupRequest{PickupDate: time.Now()})
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestPickupService_Schedule_RepoError(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewPickupService(pr, nil)
	_, err := svc.Schedule(uuid.New(), &domain.SchedulePickupRequest{PickupDate: time.Now()})
	assert.Error(t, err)
	assert.False(t, errors.Is(err, ErrNotFound))
}

func TestPickupService_Schedule_UpdateError(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{ID: id, Type: domain.PickupTypeOrganic, Status: domain.PickupStatusPending}, nil
		},
		updateFn: func(p *domain.WastePickup) error {
			return errors.New("db error")
		},
	}
	svc := NewPickupService(pr, nil)
	_, err := svc.Schedule(uuid.New(), &domain.SchedulePickupRequest{PickupDate: time.Now()})
	assert.Error(t, err)
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

func TestPickupService_Complete_NotFound(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	svc := NewPickupService(pr, nil)
	_, _, err := svc.Complete(uuid.New())
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestPickupService_Complete_RepoErrorOnFind(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewPickupService(pr, nil)
	_, _, err := svc.Complete(uuid.New())
	assert.Error(t, err)
	assert.False(t, errors.Is(err, ErrNotFound))
}

func TestPickupService_Complete_RepoErrorOnUpdate(t *testing.T) {
	pickupID := uuid.New()
	pr := &mockPickupRepo{
		findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{ID: id, Type: domain.PickupTypeOrganic, Status: domain.PickupStatusScheduled}, nil
		},
		updateFn: func(p *domain.WastePickup) error {
			return errors.New("db error")
		},
	}
	pmr := &mockPaymentRepo{}
	svc := NewPickupService(pr, pmr)
	_, _, err := svc.Complete(pickupID)
	assert.Error(t, err)
}

func TestPickupService_Complete_RepoErrorOnPaymentCreate(t *testing.T) {
	pickupID := uuid.New()
	pr := &mockPickupRepo{
		findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{ID: id, Type: domain.PickupTypePlastic, Status: domain.PickupStatusScheduled}, nil
		},
		updateFn: func(p *domain.WastePickup) error { return nil },
	}
	pmr := &mockPaymentRepo{
		createFn: func(p *domain.Payment) error {
			return errors.New("db error")
		},
	}
	svc := NewPickupService(pr, pmr)
	_, _, err := svc.Complete(pickupID)
	assert.Error(t, err)
}

func TestPickupService_Complete_InvalidPickupType(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{
				ID: id, Type: "unknown", Status: domain.PickupStatusScheduled,
			}, nil
		},
		updateFn: func(p *domain.WastePickup) error { return nil },
	}
	pmr := &mockPaymentRepo{}
	svc := NewPickupService(pr, pmr)
	_, _, err := svc.Complete(uuid.New())
	assert.Equal(t, ErrInvalidPickupType, err)
}

func TestPickupService_Complete_NilPickupDate(t *testing.T) {
	pickupID := uuid.New()
	pr := &mockPickupRepo{
		findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{
				ID: id, Type: domain.PickupTypeOrganic, Status: domain.PickupStatusScheduled,
			}, nil
		},
		updateFn: func(p *domain.WastePickup) error {
			assert.NotNil(t, p.PickupDate)
			return nil
		},
	}
	pmr := &mockPaymentRepo{
		createFn: func(p *domain.Payment) error { return nil },
	}
	svc := NewPickupService(pr, pmr)
	p, _, err := svc.Complete(pickupID)
	assert.NoError(t, err)
	assert.NotNil(t, p.PickupDate)
}

func TestPickupService_Cancel_Success(t *testing.T) {
	id := uuid.New()
	pr := &mockPickupRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{ID: uid, Status: domain.PickupStatusPending}, nil
		},
		updateFn: func(p *domain.WastePickup) error { return nil },
	}
	svc := NewPickupService(pr, nil)
	p, err := svc.Cancel(id)
	assert.NoError(t, err)
	assert.Equal(t, domain.PickupStatusCanceled, p.Status)
}

func TestPickupService_Cancel_NotFound(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	svc := NewPickupService(pr, nil)
	_, err := svc.Cancel(uuid.New())
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestPickupService_Cancel_RepoErrorOnFind(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewPickupService(pr, nil)
	_, err := svc.Cancel(uuid.New())
	assert.Error(t, err)
	assert.False(t, errors.Is(err, ErrNotFound))
}

func TestPickupService_Cancel_NotPending(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{ID: id, Status: domain.PickupStatusCompleted}, nil
		},
	}
	svc := NewPickupService(pr, nil)
	_, err := svc.Cancel(uuid.New())
	assert.Equal(t, ErrPickupNotPending, err)
}

func TestPickupService_Complete_NotScheduled(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(id uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{ID: id, Status: domain.PickupStatusPending}, nil
		},
	}
	pmr := &mockPaymentRepo{}
	svc := NewPickupService(pr, pmr)
	_, _, err := svc.Complete(uuid.New())
	assert.Equal(t, ErrPickupNotSchedulable, err)
}

func TestPickupService_Cancel_RepoErrorOnUpdate(t *testing.T) {
	id := uuid.New()
	pr := &mockPickupRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{ID: uid, Status: domain.PickupStatusPending}, nil
		},
		updateFn: func(p *domain.WastePickup) error {
			return errors.New("db error")
		},
	}
	svc := NewPickupService(pr, nil)
	_, err := svc.Cancel(id)
	assert.Error(t, err)
}

func TestPickupService_Update_Success(t *testing.T) {
	id := uuid.New()
	safety := false
	pr := &mockPickupRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{ID: uid, Status: domain.PickupStatusPending, Type: domain.PickupTypeOrganic}, nil
		},
		updateFn: func(p *domain.WastePickup) error {
			assert.Equal(t, domain.PickupTypePlastic, p.Type)
			assert.True(t, p.SafetyCheck)
			return nil
		},
	}
	svc := NewPickupService(pr, nil)
	_ = safety
	s := true
	p, err := svc.Update(id, &domain.CreatePickupRequest{Type: domain.PickupTypePlastic, SafetyCheck: &s})
	assert.NoError(t, err)
	assert.NotNil(t, p)
}

func TestPickupService_Update_NotFound(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.WastePickup, error) { return nil, gorm.ErrRecordNotFound },
	}
	svc := NewPickupService(pr, nil)
	_, err := svc.Update(uuid.New(), &domain.CreatePickupRequest{Type: domain.PickupTypeOrganic, SafetyCheck: nil})
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestPickupService_Update_FindError(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.WastePickup, error) { return nil, errors.New("db down") },
	}
	svc := NewPickupService(pr, nil)
	_, err := svc.Update(uuid.New(), &domain.CreatePickupRequest{Type: domain.PickupTypeOrganic, SafetyCheck: nil})
	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrNotFound)
}

func TestPickupService_Update_NotPending(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{ID: uid, Status: domain.PickupStatusCompleted}, nil
		},
	}
	svc := NewPickupService(pr, nil)
	_, err := svc.Update(uuid.New(), &domain.CreatePickupRequest{Type: domain.PickupTypePaper, SafetyCheck: nil})
	assert.ErrorIs(t, err, ErrPickupNotPending)
}

func TestPickupService_Update_RepoUpdateError(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{ID: uid, Status: domain.PickupStatusPending, Type: domain.PickupTypeOrganic}, nil
		},
		updateFn: func(p *domain.WastePickup) error { return errors.New("save failed") },
	}
	svc := NewPickupService(pr, nil)
	_, err := svc.Update(uuid.New(), &domain.CreatePickupRequest{Type: domain.PickupTypePlastic, SafetyCheck: nil})
	assert.Error(t, err)
}

func TestPickupService_Update_InvalidType(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{ID: uid, Status: domain.PickupStatusPending}, nil
		},
	}
	svc := NewPickupService(pr, nil)
	_, err := svc.Update(uuid.New(), &domain.CreatePickupRequest{Type: "invalid", SafetyCheck: nil})
	assert.ErrorIs(t, err, ErrInvalidPickupType)
}

func TestPickupService_Delete_Success(t *testing.T) {
	id := uuid.New()
	pr := &mockPickupRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{ID: uid, Status: domain.PickupStatusPending}, nil
		},
		deleteFn: func(uid uuid.UUID) error { return nil },
	}
	svc := NewPickupService(pr, nil)
	err := svc.Delete(id)
	assert.NoError(t, err)
}

func TestPickupService_Delete_NotFound(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.WastePickup, error) { return nil, gorm.ErrRecordNotFound },
	}
	svc := NewPickupService(pr, nil)
	err := svc.Delete(uuid.New())
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestPickupService_Delete_FindError(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.WastePickup, error) { return nil, errors.New("db down") },
	}
	svc := NewPickupService(pr, nil)
	err := svc.Delete(uuid.New())
	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrNotFound)
}

func TestPickupService_Delete_RepoError(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{ID: uid, Status: domain.PickupStatusPending}, nil
		},
		deleteFn: func(uid uuid.UUID) error { return errors.New("delete failed") },
	}
	svc := NewPickupService(pr, nil)
	err := svc.Delete(uuid.New())
	assert.Error(t, err)
}

func TestPickupService_Delete_NotPending(t *testing.T) {
	pr := &mockPickupRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.WastePickup, error) {
			return &domain.WastePickup{ID: uid, Status: domain.PickupStatusCompleted}, nil
		},
	}
	svc := NewPickupService(pr, nil)
	err := svc.Delete(uuid.New())
	assert.ErrorIs(t, err, ErrPickupNotPending)
}
