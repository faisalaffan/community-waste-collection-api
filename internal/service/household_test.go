package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
)

type mockHouseholdRepo struct {
	createFn   func(h *domain.Household) error
	findByIDFn func(id uuid.UUID) (*domain.Household, error)
	findAllFn  func(page, perPage int) ([]domain.Household, int64, error)
	deleteFn   func(id uuid.UUID) error
}

func (m *mockHouseholdRepo) Create(h *domain.Household) error                       { return m.createFn(h) }
func (m *mockHouseholdRepo) FindByID(id uuid.UUID) (*domain.Household, error)       { return m.findByIDFn(id) }
func (m *mockHouseholdRepo) FindAll(page, perPage int) ([]domain.Household, int64, error) { return m.findAllFn(page, perPage) }
func (m *mockHouseholdRepo) Delete(id uuid.UUID) error                              { return m.deleteFn(id) }

func TestHouseholdService_Create(t *testing.T) {
	repo := &mockHouseholdRepo{
		createFn: func(h *domain.Household) error { return nil },
	}
	svc := NewHouseholdService(repo)
	h, err := svc.Create(&domain.CreateHouseholdRequest{OwnerName: "Budi", Address: "Jl. A"})
	assert.NoError(t, err)
	assert.Equal(t, "Budi", h.OwnerName)
	assert.NotEqual(t, uuid.Nil, h.ID)
}

func TestHouseholdService_GetByID_NotFound(t *testing.T) {
	repo := &mockHouseholdRepo{
		findByIDFn: func(id uuid.UUID) (*domain.Household, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	svc := NewHouseholdService(repo)
	_, err := svc.GetByID(uuid.New())
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestHouseholdService_List(t *testing.T) {
	repo := &mockHouseholdRepo{
		findAllFn: func(page, perPage int) ([]domain.Household, int64, error) {
			return []domain.Household{}, int64(0), nil
		},
	}
	svc := NewHouseholdService(repo)
	list, total, err := svc.List(1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, list)
}

func TestHouseholdService_Delete(t *testing.T) {
	id := uuid.New()
	repo := &mockHouseholdRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.Household, error) {
			return &domain.Household{ID: uid}, nil
		},
		deleteFn: func(uid uuid.UUID) error { return nil },
	}
	svc := NewHouseholdService(repo)
	err := svc.Delete(id)
	assert.NoError(t, err)
}
