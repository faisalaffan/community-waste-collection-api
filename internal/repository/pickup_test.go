package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
)

func TestPickupRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPickupRepository(db)
	p := &domain.WastePickup{ID: uuid.New(), HouseholdID: uuid.New(), Type: domain.PickupTypeOrganic}
	err := repo.Create(p)
	assert.NoError(t, err)
	assert.Equal(t, domain.PickupStatusPending, p.Status)
}

func TestPickupRepo_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPickupRepository(db)
	p := &domain.WastePickup{ID: uuid.New(), HouseholdID: uuid.New(), Type: domain.PickupTypePlastic}
	repo.Create(p)
	found, err := repo.FindByID(p.ID)
	assert.NoError(t, err)
	assert.Equal(t, domain.PickupTypePlastic, found.Type)
}

func TestPickupRepo_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPickupRepository(db)
	_, err := repo.FindByID(uuid.New())
	assert.Error(t, err)
}

func TestPickupRepo_FindAll_WithFilters(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPickupRepository(db)
	hID := uuid.New()
	repo.Create(&domain.WastePickup{ID: uuid.New(), HouseholdID: hID, Type: domain.PickupTypeOrganic, Status: domain.PickupStatusPending})
	repo.Create(&domain.WastePickup{ID: uuid.New(), HouseholdID: hID, Type: domain.PickupTypePaper, Status: domain.PickupStatusScheduled})
	list, total, err := repo.FindAll(PickupFilter{HouseholdID: hID, Page: 1, PerPage: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, list, 2)
}

func TestPickupRepo_FindAll_EmptyFilter(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPickupRepository(db)
	list, total, err := repo.FindAll(PickupFilter{Page: 1, PerPage: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, list)
}

func TestPickupRepo_FindAll_StatusFilter(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPickupRepository(db)
	hID := uuid.New()
	repo.Create(&domain.WastePickup{ID: uuid.New(), HouseholdID: hID, Type: domain.PickupTypeOrganic, Status: domain.PickupStatusPending})
	repo.Create(&domain.WastePickup{ID: uuid.New(), HouseholdID: hID, Type: domain.PickupTypePaper, Status: domain.PickupStatusScheduled})
	list, total, err := repo.FindAll(PickupFilter{Status: "scheduled", Page: 1, PerPage: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
	assert.Equal(t, domain.PickupStatusScheduled, list[0].Status)
}

func TestPickupRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPickupRepository(db)
	p := &domain.WastePickup{ID: uuid.New(), HouseholdID: uuid.New(), Type: domain.PickupTypeOrganic}
	repo.Create(p)

	p.Status = domain.PickupStatusScheduled
	err := repo.Update(p)
	assert.NoError(t, err)

	found, _ := repo.FindByID(p.ID)
	assert.Equal(t, domain.PickupStatusScheduled, found.Status)
}

func TestPickupRepo_CancelOrganicPending(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPickupRepository(db)
	oldTime := time.Now().Add(-4 * 24 * time.Hour)
	p := &domain.WastePickup{
		ID: uuid.New(), HouseholdID: uuid.New(), Type: domain.PickupTypeOrganic,
		Status: domain.PickupStatusPending, CreatedAt: oldTime, UpdatedAt: oldTime,
	}
	db.Create(p)
	affected, err := repo.CancelOrganicPending(3 * 24 * time.Hour)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), affected)
}
