package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
	"github.com/faisalaffan/community-waste-collection-api/internal/repository"
)

func TestPaymentService_Create_Success(t *testing.T) {
	pmr := &mockPaymentRepo{
		createFn: func(p *domain.Payment) error { return nil },
	}
	svc := NewPaymentService(pmr, nil)
	p, err := svc.Create(&domain.CreatePaymentRequest{HouseholdID: uuid.New()})
	assert.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusPending, p.Status)
	assert.NotEqual(t, uuid.Nil, p.ID)
}

func TestPaymentService_Create_RepoError(t *testing.T) {
	pmr := &mockPaymentRepo{
		createFn: func(p *domain.Payment) error { return errors.New("db error") },
	}
	svc := NewPaymentService(pmr, nil)
	_, err := svc.Create(&domain.CreatePaymentRequest{HouseholdID: uuid.New()})
	assert.Error(t, err)
}

func TestPaymentService_GetByID_Success(t *testing.T) {
	id := uuid.New()
	pmr := &mockPaymentRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.Payment, error) {
			return &domain.Payment{ID: uid, Amount: 50000}, nil
		},
	}
	svc := NewPaymentService(pmr, nil)
	p, err := svc.GetByID(id)
	assert.NoError(t, err)
	assert.Equal(t, 50000.0, p.Amount)
}

func TestPaymentService_GetByID_NotFound(t *testing.T) {
	pmr := &mockPaymentRepo{
		findByIDFn: func(id uuid.UUID) (*domain.Payment, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	svc := NewPaymentService(pmr, nil)
	_, err := svc.GetByID(uuid.New())
	assert.Error(t, err)
}

func TestPaymentService_GetByID_RepoError(t *testing.T) {
	pmr := &mockPaymentRepo{
		findByIDFn: func(id uuid.UUID) (*domain.Payment, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewPaymentService(pmr, nil)
	_, err := svc.GetByID(uuid.New())
	assert.Error(t, err)
	assert.False(t, errors.Is(err, ErrNotFound))
}

func TestPaymentService_List(t *testing.T) {
	pmr := &mockPaymentRepo{
		findAllFn: func(filter repository.PaymentFilter) ([]domain.Payment, int64, error) {
			return []domain.Payment{}, int64(0), nil
		},
	}
	svc := NewPaymentService(pmr, nil)
	list, total, err := svc.List(repository.PaymentFilter{Page: 1, PerPage: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, list, 0)
}

func TestPaymentService_List_Defaults(t *testing.T) {
	pmr := &mockPaymentRepo{
		findAllFn: func(filter repository.PaymentFilter) ([]domain.Payment, int64, error) {
			assert.Equal(t, 1, filter.Page)
			assert.Equal(t, 10, filter.PerPage)
			return []domain.Payment{}, 0, nil
		},
	}
	svc := NewPaymentService(pmr, nil)
	svc.List(repository.PaymentFilter{Page: 0, PerPage: 0})
}
