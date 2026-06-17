package service

import (
	"errors"
	"io"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/faisalaffan/community-waste-collection-api/internal/domain"
	"github.com/faisalaffan/community-waste-collection-api/internal/repository"
)

type mockFileStorage struct {
	uploadFn   func(r io.Reader, objectName string, size int64, contentType string) (string, error)
	downloadFn func(objectName string) (io.ReadCloser, error)
}

func (m *mockFileStorage) Upload(r io.Reader, objectName string, size int64, contentType string) (string, error) {
	return m.uploadFn(r, objectName, size, contentType)
}

func (m *mockFileStorage) Download(objectName string) (io.ReadCloser, error) {
	if m.downloadFn != nil {
		return m.downloadFn(objectName)
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func TestPaymentService_Create_Success(t *testing.T) {
	pmr := &mockPaymentRepo{
		createFn: func(p *domain.Payment) error { return nil },
	}
	svc := NewPaymentService(pmr, nil)
	p, err := svc.Create(&domain.CreatePaymentRequest{HouseholdID: uuid.New()})
	assert.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusPending, p.Status)
	assert.NotEqual(t, uuid.Nil, p.ID)
}

func TestPaymentService_Create_RepoError(t *testing.T) {
	pmr := &mockPaymentRepo{
		createFn: func(p *domain.Payment) error { return errors.New("db error") },
	}
	svc := NewPaymentService(pmr, nil)
	_, err := svc.Create(&domain.CreatePaymentRequest{HouseholdID: uuid.New()})
	assert.Error(t, err)
}

func TestPaymentService_GetByID_Success(t *testing.T) {
	id := uuid.New()
	pmr := &mockPaymentRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.Payment, error) {
			return &domain.Payment{ID: uid, Amount: 50000}, nil
		},
	}
	svc := NewPaymentService(pmr, nil)
	p, err := svc.GetByID(id)
	assert.NoError(t, err)
	assert.Equal(t, 50000.0, p.Amount)
}

func TestPaymentService_GetByID_NotFound(t *testing.T) {
	pmr := &mockPaymentRepo{
		findByIDFn: func(id uuid.UUID) (*domain.Payment, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	svc := NewPaymentService(pmr, nil)
	_, err := svc.GetByID(uuid.New())
	assert.Error(t, err)
}

func TestPaymentService_GetByID_RepoError(t *testing.T) {
	pmr := &mockPaymentRepo{
		findByIDFn: func(id uuid.UUID) (*domain.Payment, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewPaymentService(pmr, nil)
	_, err := svc.GetByID(uuid.New())
	assert.Error(t, err)
	assert.False(t, errors.Is(err, ErrNotFound))
}

func TestPaymentService_List(t *testing.T) {
	pmr := &mockPaymentRepo{
		findAllFn: func(filter repository.PaymentFilter) ([]domain.Payment, int64, error) {
			return []domain.Payment{}, int64(0), nil
		},
	}
	svc := NewPaymentService(pmr, nil)
	list, total, err := svc.List(repository.PaymentFilter{Page: 1, PerPage: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, list, 0)
}

func TestPaymentService_List_Defaults(t *testing.T) {
	pmr := &mockPaymentRepo{
		findAllFn: func(filter repository.PaymentFilter) ([]domain.Payment, int64, error) {
			assert.Equal(t, 1, filter.Page)
			assert.Equal(t, 10, filter.PerPage)
			return []domain.Payment{}, 0, nil
		},
	}
	svc := NewPaymentService(pmr, nil)
	svc.List(repository.PaymentFilter{Page: 0, PerPage: 0})
}

// --- Confirm tests ---

func newTestFileHeader() *multipart.FileHeader {
	body := "--BOUNDARY\r\nContent-Disposition: form-data; name=\"proof_file\"; filename=\"proof.jpg\"\r\nContent-Type: image/jpeg\r\n\r\nfake-image-data\r\n--BOUNDARY--\r\n"
	reader := multipart.NewReader(strings.NewReader(body), "BOUNDARY")
	form, err := reader.ReadForm(1024)
	if err != nil {
		panic(err)
	}
	return form.File["proof_file"][0]
}

func TestPaymentService_Confirm_Success(t *testing.T) {
	id := uuid.New()
	pmr := &mockPaymentRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.Payment, error) {
			return &domain.Payment{ID: uid, Status: domain.PaymentStatusPending}, nil
		},
		updateFn: func(p *domain.Payment) error { return nil },
	}
	mfs := &mockFileStorage{
		uploadFn: func(r io.Reader, objectName string, size int64, contentType string) (string, error) {
			return "https://s3.example.com/proof.jpg", nil
		},
	}
	svc := NewPaymentService(pmr, mfs)
	p, err := svc.Confirm(id, newTestFileHeader())
	assert.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusPaid, p.Status)
	assert.NotNil(t, p.PaymentDate)
	assert.NotNil(t, p.ProofFileURL)
	assert.Equal(t, "https://s3.example.com/proof.jpg", *p.ProofFileURL)
}

func TestPaymentService_Confirm_FindByIDError(t *testing.T) {
	pmr := &mockPaymentRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.Payment, error) {
			return nil, errors.New("db connection lost")
		},
	}
	svc := NewPaymentService(pmr, nil)
	_, err := svc.Confirm(uuid.New(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db connection lost")
}

func TestPaymentService_Confirm_NotFound(t *testing.T) {
	pmr := &mockPaymentRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.Payment, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	svc := NewPaymentService(pmr, nil)
	_, err := svc.Confirm(uuid.New(), nil)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrPaymentNotFound))
}

func TestPaymentService_Confirm_NotPending(t *testing.T) {
	pmr := &mockPaymentRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.Payment, error) {
			return &domain.Payment{ID: uid, Status: domain.PaymentStatusPaid}, nil
		},
	}
	svc := NewPaymentService(pmr, nil)
	_, err := svc.Confirm(uuid.New(), nil)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrPaymentNotPending))
}

func TestPaymentService_Confirm_NilFile(t *testing.T) {
	id := uuid.New()
	pmr := &mockPaymentRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.Payment, error) {
			return &domain.Payment{ID: uid, Status: domain.PaymentStatusPending}, nil
		},
	}
	svc := NewPaymentService(pmr, nil)
	_, err := svc.Confirm(id, nil)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrProofFileRequired))
}

func TestPaymentService_Confirm_FileOpenError(t *testing.T) {
	id := uuid.New()
	pmr := &mockPaymentRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.Payment, error) {
			return &domain.Payment{ID: uid, Status: domain.PaymentStatusPending}, nil
		},
	}
	svc := NewPaymentService(pmr, nil)
	// A FileHeader constructed without multipart parsing has no temp file; Open() will fail.
	fh := &multipart.FileHeader{Filename: "test.jpg", Size: 100}
	_, err := svc.Confirm(id, fh)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gagal membuka file")
}

func TestPaymentService_Confirm_UploadError(t *testing.T) {
	id := uuid.New()
	pmr := &mockPaymentRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.Payment, error) {
			return &domain.Payment{ID: uid, Status: domain.PaymentStatusPending}, nil
		},
		updateFn: func(p *domain.Payment) error { return nil },
	}
	mfs := &mockFileStorage{
		uploadFn: func(r io.Reader, objectName string, size int64, contentType string) (string, error) {
			return "", errors.New("s3 connection failed")
		},
	}
	svc := NewPaymentService(pmr, mfs)
	_, err := svc.Confirm(id, newTestFileHeader())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gagal upload ke storage")
}

func TestPaymentService_Confirm_UpdateError(t *testing.T) {
	id := uuid.New()
	pmr := &mockPaymentRepo{
		findByIDFn: func(uid uuid.UUID) (*domain.Payment, error) {
			return &domain.Payment{ID: uid, Status: domain.PaymentStatusPending}, nil
		},
		updateFn: func(p *domain.Payment) error { return errors.New("db update failed") },
	}
	mfs := &mockFileStorage{
		uploadFn: func(r io.Reader, objectName string, size int64, contentType string) (string, error) {
			return "https://s3.example.com/proof.jpg", nil
		},
	}
	svc := NewPaymentService(pmr, mfs)
	_, err := svc.Confirm(id, newTestFileHeader())
	assert.Error(t, err)
}
