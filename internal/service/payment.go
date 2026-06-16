package service

import (
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
	"github.com/faisalaffan/community-waste-collection-api/internal/repository"
	"github.com/faisalaffan/community-waste-collection-api/pkg/storage"
)

var (
	ErrPaymentNotFound   = errors.New("payment tidak ditemukan")
	ErrPaymentNotPending = errors.New("hanya payment pending yang dapat dikonfirmasi")
	ErrProofFileRequired = errors.New("file bukti pembayaran wajib diupload")
)

type PaymentService interface {
	Create(req *domain.CreatePaymentRequest) (*domain.Payment, error)
	GetByID(id uuid.UUID) (*domain.Payment, error)
	List(filter repository.PaymentFilter) ([]domain.Payment, int64, error)
	Confirm(id uuid.UUID, file *multipart.FileHeader) (*domain.Payment, error)
}

type paymentService struct {
	paymentRepo repository.PaymentRepository
	storage     *storage.S3Client
}

func NewPaymentService(pr repository.PaymentRepository, s3 *storage.S3Client) PaymentService {
	return &paymentService{paymentRepo: pr, storage: s3}
}

func (s *paymentService) Create(req *domain.CreatePaymentRequest) (*domain.Payment, error) {
	p := &domain.Payment{
		ID:          uuid.New(),
		HouseholdID: req.HouseholdID,
		Amount:      0,
		Status:      domain.PaymentStatusPending,
	}
	if err := s.paymentRepo.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *paymentService) GetByID(id uuid.UUID) (*domain.Payment, error) {
	p, err := s.paymentRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *paymentService) List(filter repository.PaymentFilter) ([]domain.Payment, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 10
	}
	return s.paymentRepo.FindAll(filter)
}

// BR-06: Upload bukti pembayaran ke S3
func (s *paymentService) Confirm(id uuid.UUID, file *multipart.FileHeader) (*domain.Payment, error) {
	p, err := s.paymentRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	if p.Status != domain.PaymentStatusPending {
		return nil, ErrPaymentNotPending
	}

	if file == nil {
		return nil, ErrProofFileRequired
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("gagal membuka file: %w", err)
	}
	defer src.Close()

	ext := filepath.Ext(file.Filename)
	objectKey := fmt.Sprintf("proofs/%s/%s%s", p.ID.String(), uuid.New().String(), ext)

	url, err := s.storage.Upload(src, objectKey, file.Size, file.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("gagal upload ke storage: %w", err)
	}

	now := time.Now()
	p.Status = domain.PaymentStatusPaid
	p.PaymentDate = &now
	p.ProofFileURL = &url

	if err := s.paymentRepo.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}
