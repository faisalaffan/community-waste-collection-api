package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
)

func TestPaymentRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPaymentRepository(db)
	p := &domain.Payment{ID: uuid.New(), HouseholdID: uuid.New(), WasteID: uuid.New(), Amount: 50000}
	err := repo.Create(p)
	assert.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusPending, p.Status)
}

func TestPaymentRepo_HasPendingByHousehold(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPaymentRepository(db)
	hID := uuid.New()
	repo.Create(&domain.Payment{ID: uuid.New(), HouseholdID: hID, WasteID: uuid.New(), Amount: 50000, Status: domain.PaymentStatusPending})
	has, err := repo.HasPendingByHousehold(hID)
	assert.NoError(t, err)
	assert.True(t, has)
}
