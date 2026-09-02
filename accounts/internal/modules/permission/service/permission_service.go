package service

import (
	"context"
	"fmt"
	"strings"

	"accounts/internal/models"
	"accounts/internal/modules/permission/repository"
	"accounts/internal/pkg/apperrors"

	"github.com/google/uuid"
)

// PermissionService defines the business logic contract for permissions
type PermissionService interface {
	CreatePermission(ctx context.Context, name string) (*models.Permission, error)
	GetPermission(ctx context.Context, id uuid.UUID) (*models.Permission, error)
	UpdatePermission(ctx context.Context, id uuid.UUID, name string) (*models.Permission, error)
	DeletePermission(ctx context.Context, id uuid.UUID) error
	ListPermissions(ctx context.Context, page, pageSize int, search string) ([]models.Permission, int64, error)
}

type permissionService struct {
	repo repository.PermissionRepository
}

// NewPermissionService creates a new PermissionService instance
func NewPermissionService(repo repository.PermissionRepository) PermissionService {
	return &permissionService{repo: repo}
}

func (s *permissionService) CreatePermission(ctx context.Context, name string) (*models.Permission, error) {
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return nil, fmt.Errorf("%w: name cannot be empty", apperrors.ErrInvalidArgument)
	}

	existing, err := s.repo.FindByName(ctx, cleanName)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("%w: permission with name '%s' already exists", apperrors.ErrAlreadyExists, cleanName)
	}

	permission := &models.Permission{
		ID:   uuid.New(),
		Name: cleanName,
	}

	if err := s.repo.Create(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return permission, nil
}

func (s *permissionService) GetPermission(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid permission ID", apperrors.ErrInvalidArgument)
	}

	permission, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve permission: %w", err)
	}
	if permission == nil {
		return nil, fmt.Errorf("%w: permission not found", apperrors.ErrNotFound)
	}

	return permission, nil
}

func (s *permissionService) UpdatePermission(ctx context.Context, id uuid.UUID, name string) (*models.Permission, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid permission ID", apperrors.ErrInvalidArgument)
	}

	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return nil, fmt.Errorf("%w: name cannot be empty", apperrors.ErrInvalidArgument)
	}

	permission, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find permission: %w", err)
	}
	if permission == nil {
		return nil, fmt.Errorf("%w: permission not found", apperrors.ErrNotFound)
	}

	// Check if another permission already uses the new name
	existing, err := s.repo.FindByName(ctx, cleanName)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.ID != id {
		return nil, fmt.Errorf("%w: permission name '%s' is already in use", apperrors.ErrAlreadyExists, cleanName)
	}

	permission.Name = cleanName
	if err := s.repo.Update(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	return permission, nil
}

func (s *permissionService) DeletePermission(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("%w: invalid permission ID", apperrors.ErrInvalidArgument)
	}

	permission, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find permission: %w", err)
	}
	if permission == nil {
		return fmt.Errorf("%w: permission not found", apperrors.ErrNotFound)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	return nil
}

func (s *permissionService) ListPermissions(ctx context.Context, page, pageSize int, search string) ([]models.Permission, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.repo.List(ctx, page, pageSize, strings.TrimSpace(search))
}
