package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	// Create tables manually — SQLite does not support gen_random_uuid() DEFAULT.
	db.Exec(`CREATE TABLE households (
		id text PRIMARY KEY,
		owner_name text NOT NULL,
		address text NOT NULL,
		created_at datetime,
		updated_at datetime
	)`)
	db.Exec(`CREATE TABLE waste_pickups (
		id text PRIMARY KEY,
		household_id text NOT NULL,
		type text NOT NULL,
		status text NOT NULL DEFAULT 'pending',
		pickup_date datetime,
		safety_check integer NOT NULL DEFAULT 0,
		created_at datetime,
		updated_at datetime
	)`)
	db.Exec(`CREATE TABLE payments (
		id text PRIMARY KEY,
		household_id text NOT NULL,
		waste_id text NOT NULL,
		amount real NOT NULL,
		payment_date datetime,
		status text NOT NULL DEFAULT 'pending',
		proof_file_url text,
		created_at datetime,
		updated_at datetime
	)`)

	return db
}

func TestHouseholdRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHouseholdRepository(db)
	h := &domain.Household{ID: uuid.New(), OwnerName: "Budi", Address: "Jl. A"}
	err := repo.Create(h)
	assert.NoError(t, err)
}

func TestHouseholdRepo_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHouseholdRepository(db)
	h := &domain.Household{ID: uuid.New(), OwnerName: "Ani", Address: "Jl. B"}
	repo.Create(h)
	found, err := repo.FindByID(h.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Ani", found.OwnerName)
}

func TestHouseholdRepo_FindAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHouseholdRepository(db)
	repo.Create(&domain.Household{ID: uuid.New(), OwnerName: "A", Address: "X"})
	repo.Create(&domain.Household{ID: uuid.New(), OwnerName: "B", Address: "Y"})
	list, total, err := repo.FindAll(1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, list, 2)
}

func TestHouseholdRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHouseholdRepository(db)
	h := &domain.Household{ID: uuid.New(), OwnerName: "C", Address: "Z"}
	repo.Create(h)
	err := repo.Delete(h.ID)
	assert.NoError(t, err)
	_, err = repo.FindByID(h.ID)
	assert.Error(t, err)
}
