package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewHousehold(t *testing.T) {
	req := CreateHouseholdRequest{OwnerName: "Budi", Address: "Jl. Merdeka No.1"}
	if req.OwnerName != "Budi" {
		t.Errorf("OwnerName = %s, want Budi", req.OwnerName)
	}
}

func TestPickupAmounts(t *testing.T) {
	if PickupAmounts[PickupTypeOrganic] != 50000 {
		t.Errorf("organic = %f, want 50000", PickupAmounts[PickupTypeOrganic])
	}
	if PickupAmounts[PickupTypePlastic] != 50000 {
		t.Errorf("plastic = %f, want 50000", PickupAmounts[PickupTypePlastic])
	}
	if PickupAmounts[PickupTypePaper] != 50000 {
		t.Errorf("paper = %f, want 50000", PickupAmounts[PickupTypePaper])
	}
	if PickupAmounts[PickupTypeElectronic] != 100000 {
		t.Errorf("electronic = %f, want 100000", PickupAmounts[PickupTypeElectronic])
	}
}

func TestUUIDGeneration(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	if id1 == id2 {
		t.Error("UUIDs should be unique")
	}
}
