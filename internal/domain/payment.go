package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	PaymentStatusPending = "pending"
	PaymentStatusPaid    = "paid"
	PaymentStatusFailed  = "failed"
)

var PickupAmounts = map[string]float64{
	PickupTypeOrganic:    50000,
	PickupTypePlastic:    50000,
	PickupTypePaper:      50000,
	PickupTypeElectronic: 100000,
}

type Payment struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	HouseholdID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"household_id"`
	WasteID      uuid.UUID  `gorm:"type:uuid;not null" json:"waste_id"`
	Amount       float64    `gorm:"type:decimal(10,2);not null" json:"amount"`
	PaymentDate  *time.Time `json:"payment_date"`
	Status       string     `gorm:"type:varchar(20);default:pending" json:"status"`
	ProofFileURL *string    `json:"proof_file_url"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type CreatePaymentRequest struct {
	HouseholdID uuid.UUID `json:"household_id" validate:"required"`
}
