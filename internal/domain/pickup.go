package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	PickupTypeOrganic    = "organic"
	PickupTypePlastic    = "plastic"
	PickupTypePaper      = "paper"
	PickupTypeElectronic = "electronic"

	PickupStatusPending   = "pending"
	PickupStatusScheduled = "scheduled"
	PickupStatusCompleted = "completed"
	PickupStatusCanceled  = "canceled"
)

type WastePickup struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	HouseholdID uuid.UUID  `gorm:"type:uuid;not null;index" json:"household_id"`
	Type        string     `gorm:"type:varchar(20);not null" json:"type"`
	Status      string     `gorm:"type:varchar(20);default:pending" json:"status"`
	PickupDate  *time.Time `json:"pickup_date"`
	SafetyCheck bool       `gorm:"default:false" json:"safety_check"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreatePickupRequest struct {
	HouseholdID uuid.UUID `json:"household_id" validate:"required"`
	Type        string    `json:"type" validate:"required,oneof=organic plastic paper electronic"`
	SafetyCheck *bool     `json:"safety_check"`
}

type SchedulePickupRequest struct {
	PickupDate time.Time `json:"pickup_date" validate:"required"`
}
