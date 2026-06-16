package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
	"github.com/faisalaffan/community-waste-collection-api/internal/repository"
)

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
