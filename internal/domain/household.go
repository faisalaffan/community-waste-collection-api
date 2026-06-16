package domain

import (
	"time"

	"github.com/google/uuid"
)

type Household struct {
	ID        uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OwnerName string        `gorm:"not null" json:"owner_name"`
	Address   string        `gorm:"not null" json:"address"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Pickups   []WastePickup `gorm:"foreignKey:HouseholdID" json:"pickups,omitempty"`
	Payments  []Payment     `gorm:"foreignKey:HouseholdID" json:"payments,omitempty"`
}

type CreateHouseholdRequest struct {
	OwnerName string `json:"owner_name" validate:"required"`
	Address   string `json:"address" validate:"required"`
}
