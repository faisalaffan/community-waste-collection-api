package repository

import (
	"testing"
	"time"

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

func TestPaymentRepo_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPaymentRepository(db)
	p := &domain.Payment{ID: uuid.New(), HouseholdID: uuid.New(), WasteID: uuid.New(), Amount: 50000}
	repo.Create(p)
	found, err := repo.FindByID(p.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(50000), int64(found.Amount))
}

func TestPaymentRepo_FindAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPaymentRepository(db)
	hID := uuid.New()
	repo.Create(&domain.Payment{ID: uuid.New(), HouseholdID: hID, WasteID: uuid.New(), Amount: 50000})
	repo.Create(&domain.Payment{ID: uuid.New(), HouseholdID: hID, WasteID: uuid.New(), Amount: 100000})
	list, total, err := repo.FindAll(PaymentFilter{Page: 1, PerPage: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, list, 2)
}

func TestPaymentRepo_FindAll_WithFilters(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPaymentRepository(db)
	hID := uuid.New()
	wID := uuid.New()
	repo.Create(&domain.Payment{ID: uuid.New(), HouseholdID: hID, WasteID: wID, Amount: 50000, Status: domain.PaymentStatusPending})
	repo.Create(&domain.Payment{ID: uuid.New(), HouseholdID: hID, WasteID: wID, Amount: 100000, Status: domain.PaymentStatusPaid})
	list, total, err := repo.FindAll(PaymentFilter{HouseholdID: hID, Status: "paid", Page: 1, PerPage: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
}

func TestPaymentRepo_FindAll_WithDateFilter(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPaymentRepository(db)
	now := time.Now()
	hID := uuid.New()
	repo.Create(&domain.Payment{
		ID: uuid.New(), HouseholdID: hID, WasteID: uuid.New(),
		Amount: 50000, CreatedAt: now, UpdatedAt: now,
	})
	dayBefore := now.Add(-24 * time.Hour)
	list, total, err := repo.FindAll(PaymentFilter{
		DateFrom: &dayBefore, DateTo: &now,
		Page: 1, PerPage: 10,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
}

func TestPaymentRepo_FindAll_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPaymentRepository(db)
	list, total, err := repo.FindAll(PaymentFilter{Page: 1, PerPage: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, list)
}

func TestPaymentRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPaymentRepository(db)
	p := &domain.Payment{ID: uuid.New(), HouseholdID: uuid.New(), WasteID: uuid.New(), Amount: 50000}
	repo.Create(p)

	now := time.Now()
	p.Status = domain.PaymentStatusPaid
	p.PaymentDate = &now
	err := repo.Update(p)
	assert.NoError(t, err)

	found, _ := repo.FindByID(p.ID)
	assert.Equal(t, domain.PaymentStatusPaid, found.Status)
	assert.NotNil(t, found.PaymentDate)
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

func TestPaymentRepo_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPaymentRepository(db)
	_, err := repo.FindByID(uuid.New())
	assert.Error(t, err)
}

func TestPaymentRepo_HasPendingByHousehold_False(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPaymentRepository(db)
	hID := uuid.New()
	has, err := repo.HasPendingByHousehold(hID)
	assert.NoError(t, err)
	assert.False(t, has)
}
