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

func TestReportService_WasteSummary(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
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

	svc := NewReportService(db)
	result, err := svc.WasteSummary()
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestReportService_PaymentSummary(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
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

	svc := NewReportService(db)
	results, total, err := svc.PaymentSummary()
	assert.NoError(t, err)
	assert.Empty(t, results)
	assert.Equal(t, 0.0, total)
}

func TestReportService_PaymentSummary_WithData(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
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

	now := time.Now()
	db.Create(&domain.Payment{
		ID: uuid.New(), HouseholdID: uuid.New(), WasteID: uuid.New(),
		Amount: 50000, Status: domain.PaymentStatusPaid, CreatedAt: now, UpdatedAt: now,
	})
	db.Create(&domain.Payment{
		ID: uuid.New(), HouseholdID: uuid.New(), WasteID: uuid.New(),
		Amount: 100000, Status: domain.PaymentStatusPaid, CreatedAt: now, UpdatedAt: now,
	})
	db.Create(&domain.Payment{
		ID: uuid.New(), HouseholdID: uuid.New(), WasteID: uuid.New(),
		Amount: 50000, Status: domain.PaymentStatusPending, CreatedAt: now, UpdatedAt: now,
	})

	svc := NewReportService(db)
	results, total, err := svc.PaymentSummary()
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, 150000.0, total)
}

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

func TestReportService_PaymentSummary_DBError(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	svc := NewReportService(db)
	_, _, err := svc.PaymentSummary()
	assert.Error(t, err)
}

func TestReportService_AllHistory(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.Exec(`CREATE TABLE waste_pickups (
		id TEXT PRIMARY KEY, household_id TEXT, type TEXT, status TEXT DEFAULT 'pending',
		pickup_date DATETIME, safety_check INTEGER DEFAULT 0, created_at DATETIME, updated_at DATETIME
	)`)
	db.Exec(`CREATE TABLE payments (
		id TEXT PRIMARY KEY, household_id TEXT, waste_id TEXT, amount REAL,
		payment_date DATETIME, status TEXT DEFAULT 'pending', proof_file_url TEXT, created_at DATETIME, updated_at DATETIME
	)`)
	db.Exec(`CREATE TABLE households (id TEXT PRIMARY KEY, owner_name TEXT, address TEXT, created_at DATETIME, updated_at DATETIME)`)
	db.Exec(`INSERT INTO waste_pickups VALUES (?,?,?,?,?,?,datetime('now'),datetime('now'))`, uuid.New().String(), uuid.New().String(), "organic", "pending", nil, 0)
	db.Exec(`INSERT INTO payments VALUES (?,?,?,?,?,?,?,datetime('now'),datetime('now'))`, uuid.New().String(), uuid.New().String(), uuid.New().String(), 50000.0, nil, "pending", nil)

	svc := NewReportService(db)
	result, err := svc.AllHistory()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Pickups, 1)
	assert.Len(t, result.Payments, 1)
}

func TestReportService_HouseholdHistory_NotFound(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.Exec(`CREATE TABLE households (
		id TEXT PRIMARY KEY,
		owner_name TEXT NOT NULL,
		address TEXT NOT NULL,
		created_at DATETIME,
		updated_at DATETIME
	)`)

	svc := NewReportService(db)
	_, err := svc.HouseholdHistory(uuid.New())
	assert.Error(t, err)
	assert.True(t, IsRecordNotFound(err))
}

// IsRecordNotFound checks if the error is a gorm.ErrRecordNotFound
func IsRecordNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}
