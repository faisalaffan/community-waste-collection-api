package handler

import (
	"io"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/faisalaffan/community-waste-collection-api/internal/repository"
	"github.com/faisalaffan/community-waste-collection-api/pkg/response"
	"github.com/faisalaffan/community-waste-collection-api/pkg/storage"
)

type FileHandler struct {
	paymentRepo repository.PaymentRepository
	storage     storage.FileStorage
}

func NewFileHandler(pr repository.PaymentRepository, s3 storage.FileStorage) *FileHandler {
	return &FileHandler{paymentRepo: pr, storage: s3}
}

// Proof serves payment proof files by proxying from S3.
func (h *FileHandler) Proof(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("paymentID"))
	if err != nil {
		return response.Error(c, 400, "INVALID_ID", "id harus UUID valid")
	}

	payment, err := h.paymentRepo.FindByID(id)
	if err != nil {
		return response.Error(c, 404, "NOT_FOUND", "payment tidak ditemukan")
	}
	if payment.ProofFileURL == nil || *payment.ProofFileURL == "" {
		return response.Error(c, 404, "NOT_FOUND", "bukti pembayaran tidak tersedia")
	}

	// Extract object key from stored URL (format: http://host/bucket/objectKey)
	u, err := url.Parse(*payment.ProofFileURL)
	if err != nil {
		return response.Error(c, 500, "INTERNAL_ERROR", "gagal parse proof URL")
	}
	objectKey := strings.TrimPrefix(u.Path, "/"+strings.SplitN(u.Path, "/", 3)[1]+"/")

	reader, err := h.storage.Download(objectKey)
	if err != nil {
		return response.Error(c, 500, "INTERNAL_ERROR", "gagal download bukti pembayaran")
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return response.Error(c, 500, "INTERNAL_ERROR", "gagal membaca file")
	}

	// Detect content type from file extension
	contentType := "image/png"
	if strings.HasSuffix(objectKey, ".jpg") || strings.HasSuffix(objectKey, ".jpeg") {
		contentType = "image/jpeg"
	} else if strings.HasSuffix(objectKey, ".pdf") {
		contentType = "application/pdf"
	}

	c.Type(contentType)
	return c.Send(data)
}
