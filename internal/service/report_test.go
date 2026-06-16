package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
)

func TestReportService_HouseholdHistory(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.Exec(`CREATE TABLE households (
		id TEXT PRIMARY KEY,
		owner_name TEXT NOT NULL,
		address TEXT NOT NULL,
		created_at DATETIME,
		updated_at DATETIME
	)`)
	db.Exec(`CREATE TABLE waste_pickups (
		id TEXT PRIMARY KEY,
		household_id TEXT NOT NULL,
		type TEXT NOT NULL,
		status TEXT DEFAULT 'pending',
		pickup_date DATETIME,
		safety_check INTEGER DEFAULT 0,
		created_at DATETIME,
		updated_at DATETIME
	)`)
	db.Exec(`CREATE TABLE payments (
		id TEXT PRIMARY KEY,
		household_id TEXT NOT NULL,
		waste_id TEXT NOT NULL,
		amount REAL NOT NULL,
		payment_date DATETIME,
		status TEXT DEFAULT 'pending',
		proof_file_url TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`)

	hID := uuid.New()
	now := time.Now()
	db.Create(&domain.Household{ID: hID, OwnerName: "Test", Address: "Addr", CreatedAt: now, UpdatedAt: now})

	svc := NewReportService(db)
	history, err := svc.HouseholdHistory(hID)
	assert.NoError(t, err)
	assert.Equal(t, "Test", history.Household.OwnerName)
	assert.Len(t, history.Pickups, 0)
	assert.Len(t, history.Payments, 0)
}
