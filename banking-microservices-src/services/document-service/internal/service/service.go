// Package service implements document-service use cases: it issues presigned S3
// URLs for direct client upload/download and tracks document metadata.
package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	pkgs3 "banking-platform/pkg/s3"
	"banking-platform/services/document-service/internal/domain"
	"banking-platform/services/document-service/internal/repository"
)

// presignTTL bounds how long an upload/download URL is valid.
const presignTTL = 15 * time.Minute

// Service holds document-service dependencies. storage is nil when S3 is
// disabled, in which case presign operations are unavailable.
type Service struct {
	repo    repository.Repository
	storage *pkgs3.Client
	log     *zap.Logger
}

// New builds the Service. storage may be nil (S3 disabled).
func New(repo repository.Repository, storage *pkgs3.Client, log *zap.Logger) *Service {
	return &Service{repo: repo, storage: storage, log: log}
}

// UploadTicket is returned when a client requests to upload a document: the
// created metadata plus a presigned URL to PUT the bytes directly to S3.
type UploadTicket struct {
	Document  domain.Document `json:"document"`
	UploadURL string          `json:"upload_url,omitempty"`
}

// RequestUpload registers a document and returns a presigned upload URL.
func (s *Service) RequestUpload(ctx context.Context, ownerID, docType, contentType string) (UploadTicket, error) {
	if docType == "" {
		return UploadTicket{}, apierror.ErrBadRequest("type is required")
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	id := uuid.NewString()
	doc := domain.Document{
		ID:          id,
		OwnerID:     ownerID,
		Type:        docType,
		Key:         fmt.Sprintf("%s/%s", ownerID, id),
		ContentType: contentType,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, doc); err != nil {
		return UploadTicket{}, err
	}

	ticket := UploadTicket{Document: doc}
	if s.storage != nil {
		url, err := s.storage.PresignPut(ctx, doc.Key, contentType, presignTTL)
		if err != nil {
			return UploadTicket{}, err
		}
		ticket.UploadURL = url
	}
	return ticket, nil
}

// DownloadURL returns a presigned download URL for a document the caller owns.
func (s *Service) DownloadURL(ctx context.Context, ownerID, id string) (string, error) {
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}
	if doc.OwnerID != ownerID {
		return "", apierror.ErrForbidden("not your document")
	}
	if s.storage == nil {
		return "", apierror.New(http.StatusServiceUnavailable, "storage_disabled", "object storage is not configured")
	}
	return s.storage.PresignGet(ctx, doc.Key, presignTTL)
}

// List returns the caller's documents.
func (s *Service) List(ctx context.Context, ownerID string) ([]domain.Document, error) {
	return s.repo.ListByOwner(ctx, ownerID)
}

// Delete removes a document (metadata + object) owned by the caller.
func (s *Service) Delete(ctx context.Context, ownerID, id string) error {
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if doc.OwnerID != ownerID {
		return apierror.ErrForbidden("not your document")
	}
	if s.storage != nil {
		if err := s.storage.Delete(ctx, doc.Key); err != nil {
			return err
		}
	}
	return s.repo.Delete(ctx, id)
}
