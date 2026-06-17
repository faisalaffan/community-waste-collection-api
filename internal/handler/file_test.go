package handler

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
	"github.com/faisalaffan/community-waste-collection-api/internal/repository"
)

type mockPaymentRepoForFile struct {
	findByIDFn func(id uuid.UUID) (*domain.Payment, error)
}

func (m *mockPaymentRepoForFile) Create(p *domain.Payment) error          { return nil }
func (m *mockPaymentRepoForFile) FindByID(id uuid.UUID) (*domain.Payment, error) {
	return m.findByIDFn(id)
}
func (m *mockPaymentRepoForFile) FindAll(filter repository.PaymentFilter) ([]domain.Payment, int64, error) {
	return nil, 0, nil
}
func (m *mockPaymentRepoForFile) Update(p *domain.Payment) error { return nil }
func (m *mockPaymentRepoForFile) HasPendingByHousehold(householdID uuid.UUID) (bool, error) {
	return false, nil
}

type mockFileStorage struct {
	downloadFn func(objectName string) (io.ReadCloser, error)
	uploadFn   func(r io.Reader, objectName string, size int64, contentType string) (string, error)
}

func (m *mockFileStorage) Upload(r io.Reader, objectName string, size int64, contentType string) (string, error) {
	if m.uploadFn != nil {
		return m.uploadFn(r, objectName, size, contentType)
	}
	return "", nil
}
func (m *mockFileStorage) Download(objectName string) (io.ReadCloser, error) {
	if m.downloadFn != nil {
		return m.downloadFn(objectName)
	}
	return io.NopCloser(strings.NewReader("fake-image-data")), nil
}

func TestNewFileHandler(t *testing.T) {
	fh := NewFileHandler(nil, nil)
	assert.NotNil(t, fh)
}

func TestFileHandler_Proof_InvalidID(t *testing.T) {
	app := fiber.New()
	fh := NewFileHandler(nil, nil)
	app.Get("/files/proof/:paymentID", fh.Proof)
	req := httptest.NewRequest("GET", "/files/proof/not-uuid", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestFileHandler_Proof_NotFound(t *testing.T) {
	pr := &mockPaymentRepoForFile{
		findByIDFn: func(id uuid.UUID) (*domain.Payment, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	app := fiber.New()
	fh := NewFileHandler(pr, nil)
	app.Get("/files/proof/:paymentID", fh.Proof)
	req := httptest.NewRequest("GET", "/files/proof/"+uuid.New().String(), nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestFileHandler_Proof_NoProofFile(t *testing.T) {
	pr := &mockPaymentRepoForFile{
		findByIDFn: func(id uuid.UUID) (*domain.Payment, error) {
			return &domain.Payment{ID: id, Status: domain.PaymentStatusPaid}, nil
		},
	}
	app := fiber.New()
	fh := NewFileHandler(pr, nil)
	app.Get("/files/proof/:paymentID", fh.Proof)
	req := httptest.NewRequest("GET", "/files/proof/"+uuid.New().String(), nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestFileHandler_Proof_Success(t *testing.T) {
	url := "http://s3.example.com/bucket/proofs/some-id/file.png"
	pr := &mockPaymentRepoForFile{
		findByIDFn: func(id uuid.UUID) (*domain.Payment, error) {
			return &domain.Payment{ID: id, ProofFileURL: &url, Status: domain.PaymentStatusPaid}, nil
		},
	}
	fs := &mockFileStorage{}
	app := fiber.New()
	fh := NewFileHandler(pr, fs)
	app.Get("/files/proof/:paymentID", fh.Proof)
	req := httptest.NewRequest("GET", "/files/proof/"+uuid.New().String(), nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestFileHandler_Proof_SuccessJPG(t *testing.T) {
	url := "http://s3.example.com/bucket/proofs/some-id/file.jpeg"
	pr := &mockPaymentRepoForFile{
		findByIDFn: func(id uuid.UUID) (*domain.Payment, error) {
			return &domain.Payment{ID: id, ProofFileURL: &url, Status: domain.PaymentStatusPaid}, nil
		},
	}
	fs := &mockFileStorage{}
	app := fiber.New()
	fh := NewFileHandler(pr, fs)
	app.Get("/files/proof/:paymentID", fh.Proof)
	req := httptest.NewRequest("GET", "/files/proof/"+uuid.New().String(), nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestFileHandler_Proof_SuccessPDF(t *testing.T) {
	url := "http://s3.example.com/bucket/proofs/some-id/file.pdf"
	pr := &mockPaymentRepoForFile{
		findByIDFn: func(id uuid.UUID) (*domain.Payment, error) {
			return &domain.Payment{ID: id, ProofFileURL: &url, Status: domain.PaymentStatusPaid}, nil
		},
	}
	fs := &mockFileStorage{}
	app := fiber.New()
	fh := NewFileHandler(pr, fs)
	app.Get("/files/proof/:paymentID", fh.Proof)
	req := httptest.NewRequest("GET", "/files/proof/"+uuid.New().String(), nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestFileHandler_Proof_DownloadError(t *testing.T) {
	url := "http://s3.example.com/bucket/proofs/x/file.png"
	pr := &mockPaymentRepoForFile{
		findByIDFn: func(id uuid.UUID) (*domain.Payment, error) {
			return &domain.Payment{ID: id, ProofFileURL: &url, Status: domain.PaymentStatusPaid}, nil
		},
	}
	fs := &mockFileStorage{
		downloadFn: func(objectName string) (io.ReadCloser, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	app := fiber.New()
	fh := NewFileHandler(pr, fs)
	app.Get("/files/proof/:paymentID", fh.Proof)
	req := httptest.NewRequest("GET", "/files/proof/"+uuid.New().String(), nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 500, resp.StatusCode)
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (int, error) { return 0, gorm.ErrRecordNotFound }
func (e *errorReader) Close() error               { return nil }

func TestFileHandler_Proof_ReadError(t *testing.T) {
	url := "http://s3.example.com/bucket/proofs/x/file.png"
	pr := &mockPaymentRepoForFile{
		findByIDFn: func(id uuid.UUID) (*domain.Payment, error) {
			return &domain.Payment{ID: id, ProofFileURL: &url, Status: domain.PaymentStatusPaid}, nil
		},
	}
	fs := &mockFileStorage{
		downloadFn: func(objectName string) (io.ReadCloser, error) {
			return &errorReader{}, nil
		},
	}
	app := fiber.New()
	fh := NewFileHandler(pr, fs)
	app.Get("/files/proof/:paymentID", fh.Proof)
	req := httptest.NewRequest("GET", "/files/proof/"+uuid.New().String(), nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 500, resp.StatusCode)
}

func TestFileHandler_Proof_InvalidURL(t *testing.T) {
	badURL := "://invalid-url"
	pr := &mockPaymentRepoForFile{
		findByIDFn: func(id uuid.UUID) (*domain.Payment, error) {
			return &domain.Payment{ID: id, ProofFileURL: &badURL, Status: domain.PaymentStatusPaid}, nil
		},
	}
	app := fiber.New()
	fh := NewFileHandler(pr, nil)
	app.Get("/files/proof/:paymentID", fh.Proof)
	req := httptest.NewRequest("GET", "/files/proof/"+uuid.New().String(), nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 500, resp.StatusCode)
}
