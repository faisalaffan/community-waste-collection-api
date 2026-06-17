package service

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
	"github.com/faisalaffan/community-waste-collection-api/internal/repository"
)

var ErrNotFound = errors.New("resource not found")

type HouseholdService interface {
	Create(req *domain.CreateHouseholdRequest) (*domain.Household, error)
	GetByID(id uuid.UUID) (*domain.Household, error)
	List(page, perPage int) ([]domain.Household, int64, error)
	Update(id uuid.UUID, req *domain.CreateHouseholdRequest) (*domain.Household, error)
	Delete(id uuid.UUID) error
}

type householdService struct {
	repo repository.HouseholdRepository
}

func NewHouseholdService(repo repository.HouseholdRepository) HouseholdService {
	return &householdService{repo: repo}
}

func (s *householdService) Create(req *domain.CreateHouseholdRequest) (*domain.Household, error) {
	h := &domain.Household{
		ID:        uuid.New(),
		OwnerName: req.OwnerName,
		Address:   req.Address,
	}
	if err := s.repo.Create(h); err != nil {
		return nil, err
	}
	return h, nil
}

func (s *householdService) GetByID(id uuid.UUID) (*domain.Household, error) {
	h, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return h, nil
}

func (s *householdService) List(page, perPage int) ([]domain.Household, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}
	return s.repo.FindAll(page, perPage)
}

func (s *householdService) Update(id uuid.UUID, req *domain.CreateHouseholdRequest) (*domain.Household, error) {
	h, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	h.OwnerName = req.OwnerName
	h.Address = req.Address
	if err := s.repo.Update(h); err != nil {
		return nil, err
	}
	return h, nil
}

func (s *householdService) Delete(id uuid.UUID) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}
