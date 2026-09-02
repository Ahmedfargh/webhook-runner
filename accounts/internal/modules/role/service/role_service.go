package service

import (
	"context"
	"fmt"
	"strings"

	"accounts/internal/models"
	permissionRepo "accounts/internal/modules/permission/repository"
	"accounts/internal/modules/role/repository"
	"accounts/internal/pkg/apperrors"

	"github.com/google/uuid"
)

// RoleService defines business logic for roles
type RoleService interface {
	CreateRole(ctx context.Context, name string, permissionIDs []string) (*models.Role, error)
	GetRole(ctx context.Context, id uuid.UUID) (*models.Role, error)
	UpdateRole(ctx context.Context, id uuid.UUID, name string, permissionIDs []string) (*models.Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
	ListRoles(ctx context.Context, page, pageSize int, search string) ([]models.Role, int64, error)
	AssignPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []string) (*models.Role, error)
}

type roleService struct {
	roleRepo       repository.RoleRepository
	permissionRepo permissionRepo.PermissionRepository
}

// NewRoleService creates a new RoleService with dependencies
func NewRoleService(roleRepo repository.RoleRepository, permRepo permissionRepo.PermissionRepository) RoleService {
	return &roleService{
		roleRepo:       roleRepo,
		permissionRepo: permRepo,
	}
}

func (s *roleService) parseAndValidatePermissionIDs(ctx context.Context, ids []string) ([]models.Permission, error) {
	if len(ids) == 0 {
		return []models.Permission{}, nil
	}

	parsedUUIDs := make([]uuid.UUID, 0, len(ids))
	for _, idStr := range ids {
		id, err := uuid.Parse(strings.TrimSpace(idStr))
		if err != nil {
			return nil, fmt.Errorf("%w: invalid permission ID '%s'", apperrors.ErrInvalidArgument, idStr)
		}
		parsedUUIDs = append(parsedUUIDs, id)
	}

	permissions, err := s.permissionRepo.FindByIDs(ctx, parsedUUIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to verify permissions: %w", err)
	}

	if len(permissions) != len(parsedUUIDs) {
		return nil, fmt.Errorf("%w: one or more permission IDs do not exist", apperrors.ErrInvalidArgument)
	}

	return permissions, nil
}

func (s *roleService) CreateRole(ctx context.Context, name string, permissionIDs []string) (*models.Role, error) {
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return nil, fmt.Errorf("%w: role name cannot be empty", apperrors.ErrInvalidArgument)
	}

	existing, err := s.roleRepo.FindByName(ctx, cleanName)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("%w: role with name '%s' already exists", apperrors.ErrAlreadyExists, cleanName)
	}

	permissions, err := s.parseAndValidatePermissionIDs(ctx, permissionIDs)
	if err != nil {
		return nil, err
	}

	role := &models.Role{
		ID:          uuid.New(),
		Name:        cleanName,
		Permissions: permissions,
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

func (s *roleService) GetRole(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid role ID", apperrors.ErrInvalidArgument)
	}

	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find role: %w", err)
	}
	if role == nil {
		return nil, fmt.Errorf("%w: role not found", apperrors.ErrNotFound)
	}

	return role, nil
}

func (s *roleService) UpdateRole(ctx context.Context, id uuid.UUID, name string, permissionIDs []string) (*models.Role, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid role ID", apperrors.ErrInvalidArgument)
	}

	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return nil, fmt.Errorf("%w: role name cannot be empty", apperrors.ErrInvalidArgument)
	}

	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find role: %w", err)
	}
	if role == nil {
		return nil, fmt.Errorf("%w: role not found", apperrors.ErrNotFound)
	}

	existing, err := s.roleRepo.FindByName(ctx, cleanName)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.ID != id {
		return nil, fmt.Errorf("%w: role name '%s' is already in use", apperrors.ErrAlreadyExists, cleanName)
	}

	role.Name = cleanName
	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	if permissionIDs != nil {
		permissions, err := s.parseAndValidatePermissionIDs(ctx, permissionIDs)
		if err != nil {
			return nil, err
		}
		if err := s.roleRepo.ReplacePermissions(ctx, role, permissions); err != nil {
			return nil, fmt.Errorf("failed to update role permissions: %w", err)
		}
		role.Permissions = permissions
	}

	return role, nil
}

func (s *roleService) DeleteRole(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("%w: invalid role ID", apperrors.ErrInvalidArgument)
	}

	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find role: %w", err)
	}
	if role == nil {
		return fmt.Errorf("%w: role not found", apperrors.ErrNotFound)
	}

	if err := s.roleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

func (s *roleService) ListRoles(ctx context.Context, page, pageSize int, search string) ([]models.Role, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.roleRepo.List(ctx, page, pageSize, strings.TrimSpace(search))
}

func (s *roleService) AssignPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []string) (*models.Role, error) {
	if roleID == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid role ID", apperrors.ErrInvalidArgument)
	}

	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to find role: %w", err)
	}
	if role == nil {
		return nil, fmt.Errorf("%w: role not found", apperrors.ErrNotFound)
	}

	permissions, err := s.parseAndValidatePermissionIDs(ctx, permissionIDs)
	if err != nil {
		return nil, err
	}

	if err := s.roleRepo.ReplacePermissions(ctx, role, permissions); err != nil {
		return nil, fmt.Errorf("failed to assign permissions: %w", err)
	}

	role.Permissions = permissions
	return role, nil
}
