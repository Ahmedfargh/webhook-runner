package service

import (
	"context"
	"errors"
	"time"

	"auditService/internal/models"
	"auditService/internal/modules/audit/repository"

	"github.com/google/uuid"
)

type AuditService interface {
	RecordLog(ctx context.Context, log *models.AuditLog) (*models.AuditLog, error)
	ListLogs(ctx context.Context, filter repository.AuditFilter) ([]models.AuditLog, int64, error)
	GetLog(ctx context.Context, id uuid.UUID) (*models.AuditLog, error)
}

type auditService struct {
	repo repository.AuditRepository
}

func NewAuditService(repo repository.AuditRepository) AuditService {
	return &auditService{repo: repo}
}

func (s *auditService) RecordLog(ctx context.Context, log *models.AuditLog) (*models.AuditLog, error) {
	if log == nil {
		return nil, errors.New("audit log is nil")
	}
	if log.ServiceName == "" {
		log.ServiceName = "unknown"
	}
	if log.Action == "" {
		return nil, errors.New("action is required")
	}
	if log.Resource == "" {
		return nil, errors.New("resource is required")
	}
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now().UTC()
	}
	if log.Status == "" {
		log.Status = "SUCCESS"
	}

	if err := s.repo.Create(log); err != nil {
		return nil, err
	}
	return log, nil
}

func (s *auditService) ListLogs(ctx context.Context, filter repository.AuditFilter) ([]models.AuditLog, int64, error) {
	return s.repo.List(filter)
}

func (s *auditService) GetLog(ctx context.Context, id uuid.UUID) (*models.AuditLog, error) {
	return s.repo.FindByID(id)
}
