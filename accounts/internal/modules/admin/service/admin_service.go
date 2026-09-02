package service

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"accounts/internal/helpers/passwords"
	"accounts/internal/helpers/phonenumbers"
	"accounts/internal/models"
	"accounts/internal/modules/admin/repository"
	permRepo "accounts/internal/modules/permission/repository"
	roleRepo "accounts/internal/modules/role/repository"
	"accounts/internal/pkg/apperrors"
	countryRepo "accounts/internal/repository"

	"github.com/google/uuid"
)

// CreateAdminInput carries data required to create an admin
type CreateAdminInput struct {
	Name          string
	Email         string
	Phone         string
	Password      string
	CountryID     string
	RoleIDs       []string
	PermissionIDs []string
}

// UpdateAdminInput carries data required to update an admin
type UpdateAdminInput struct {
	ID            uuid.UUID
	Name          string
	Email         string
	Phone         string
	Password      *string
	CountryID     string
	RoleIDs       []string
	PermissionIDs []string
}

// AdminService defines business logic for admins
type AdminService interface {
	CreateAdmin(ctx context.Context, input CreateAdminInput) (*models.Admin, error)
	GetAdmin(ctx context.Context, id uuid.UUID) (*models.Admin, error)
	UpdateAdmin(ctx context.Context, input UpdateAdminInput) (*models.Admin, error)
	DeleteAdmin(ctx context.Context, id uuid.UUID) error
	ListAdmins(ctx context.Context, page, pageSize int, search string) ([]models.Admin, int64, error)
	AssignRoles(ctx context.Context, adminID uuid.UUID, roleIDs []string) (*models.Admin, error)
	AssignPermissions(ctx context.Context, adminID uuid.UUID, permissionIDs []string) (*models.Admin, error)
}

type adminService struct {
	adminRepo   repository.AdminRepository
	countryRepo countryRepo.CountryRepository
	roleRepo    roleRepo.RoleRepository
	permRepo    permRepo.PermissionRepository
}

// NewAdminService creates a new AdminService instance
func NewAdminService(
	adminRepo repository.AdminRepository,
	countryRepo countryRepo.CountryRepository,
	roleRepo roleRepo.RoleRepository,
	permRepo permRepo.PermissionRepository,
) AdminService {
	return &adminService{
		adminRepo:   adminRepo,
		countryRepo: countryRepo,
		roleRepo:    roleRepo,
		permRepo:    permRepo,
	}
}

func (s *adminService) validateEmail(email string) (string, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	if cleanEmail == "" {
		return "", fmt.Errorf("%w: email cannot be empty", apperrors.ErrInvalidArgument)
	}

	addr, err := mail.ParseAddress(cleanEmail)
	if err != nil || addr.Address != cleanEmail {
		return "", fmt.Errorf("%w: invalid email format", apperrors.ErrInvalidArgument)
	}

	return cleanEmail, nil
}

func (s *adminService) parseRoles(ctx context.Context, roleIDs []string) ([]models.Role, error) {
	if len(roleIDs) == 0 {
		return []models.Role{}, nil
	}

	uuids := make([]uuid.UUID, 0, len(roleIDs))
	for _, idStr := range roleIDs {
		id, err := uuid.Parse(strings.TrimSpace(idStr))
		if err != nil {
			return nil, fmt.Errorf("%w: invalid role ID '%s'", apperrors.ErrInvalidArgument, idStr)
		}
		uuids = append(uuids, id)
	}

	roles, err := s.roleRepo.FindByIDs(ctx, uuids)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve roles: %w", err)
	}
	if len(roles) != len(uuids) {
		return nil, fmt.Errorf("%w: one or more role IDs do not exist", apperrors.ErrInvalidArgument)
	}

	return roles, nil
}

func (s *adminService) parsePermissions(ctx context.Context, permIDs []string) ([]models.Permission, error) {
	if len(permIDs) == 0 {
		return []models.Permission{}, nil
	}

	uuids := make([]uuid.UUID, 0, len(permIDs))
	for _, idStr := range permIDs {
		id, err := uuid.Parse(strings.TrimSpace(idStr))
		if err != nil {
			return nil, fmt.Errorf("%w: invalid permission ID '%s'", apperrors.ErrInvalidArgument, idStr)
		}
		uuids = append(uuids, id)
	}

	perms, err := s.permRepo.FindByIDs(ctx, uuids)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve permissions: %w", err)
	}
	if len(perms) != len(uuids) {
		return nil, fmt.Errorf("%w: one or more permission IDs do not exist", apperrors.ErrInvalidArgument)
	}

	return perms, nil
}

func (s *adminService) CreateAdmin(ctx context.Context, input CreateAdminInput) (*models.Admin, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name cannot be empty", apperrors.ErrInvalidArgument)
	}

	email, err := s.validateEmail(input.Email)
	if err != nil {
		return nil, err
	}

	if len(input.Password) < 6 {
		return nil, fmt.Errorf("%w: password must be at least 6 characters", apperrors.ErrInvalidArgument)
	}

	countryID, err := uuid.Parse(strings.TrimSpace(input.CountryID))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid country ID format", apperrors.ErrInvalidArgument)
	}

	country, err := s.countryRepo.FindByID(ctx, countryID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify country: %w", err)
	}
	if country == nil {
		return nil, fmt.Errorf("%w: country not found", apperrors.ErrCountryNotFound)
	}

	normalizedPhone, err := phonenumbers.NormalizePhoneNumber(input.Phone, country.CountryCode)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperrors.ErrPhoneInvalid, err)
	}

	existing, err := s.adminRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("%w: email '%s' is already taken", apperrors.ErrEmailAlreadyUsed, email)
	}

	roles, err := s.parseRoles(ctx, input.RoleIDs)
	if err != nil {
		return nil, err
	}

	permissions, err := s.parsePermissions(ctx, input.PermissionIDs)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := passwords.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	admin := &models.Admin{
		ID:          uuid.New(),
		Name:        name,
		Email:       email,
		Phone:       normalizedPhone,
		Password:    hashedPassword,
		CountryID:   countryID,
		Country:     *country,
		Roles:       roles,
		Permissions: permissions,
	}

	if err := s.adminRepo.Create(ctx, admin); err != nil {
		return nil, fmt.Errorf("failed to create admin: %w", err)
	}

	return admin, nil
}

func (s *adminService) GetAdmin(ctx context.Context, id uuid.UUID) (*models.Admin, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid admin ID", apperrors.ErrInvalidArgument)
	}

	admin, err := s.adminRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find admin: %w", err)
	}
	if admin == nil {
		return nil, fmt.Errorf("%w: admin not found", apperrors.ErrNotFound)
	}

	return admin, nil
}

func (s *adminService) UpdateAdmin(ctx context.Context, input UpdateAdminInput) (*models.Admin, error) {
	if input.ID == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid admin ID", apperrors.ErrInvalidArgument)
	}

	admin, err := s.adminRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find admin: %w", err)
	}
	if admin == nil {
		return nil, fmt.Errorf("%w: admin not found", apperrors.ErrNotFound)
	}

	if input.Name != "" {
		admin.Name = strings.TrimSpace(input.Name)
	}

	if input.Email != "" {
		email, err := s.validateEmail(input.Email)
		if err != nil {
			return nil, err
		}

		existing, err := s.adminRepo.FindByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != admin.ID {
			return nil, fmt.Errorf("%w: email '%s' is already taken", apperrors.ErrEmailAlreadyUsed, email)
		}
		admin.Email = email
	}

	if input.CountryID != "" {
		countryID, err := uuid.Parse(strings.TrimSpace(input.CountryID))
		if err != nil {
			return nil, fmt.Errorf("%w: invalid country ID format", apperrors.ErrInvalidArgument)
		}
		country, err := s.countryRepo.FindByID(ctx, countryID)
		if err != nil {
			return nil, fmt.Errorf("failed to verify country: %w", err)
		}
		if country == nil {
			return nil, fmt.Errorf("%w: country not found", apperrors.ErrCountryNotFound)
		}
		admin.CountryID = countryID
		admin.Country = *country
	}

	if input.Phone != "" {
		normalizedPhone, err := phonenumbers.NormalizePhoneNumber(input.Phone, admin.Country.CountryCode)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", apperrors.ErrPhoneInvalid, err)
		}
		admin.Phone = normalizedPhone
	}

	if input.Password != nil && *input.Password != "" {
		if len(*input.Password) < 6 {
			return nil, fmt.Errorf("%w: password must be at least 6 characters", apperrors.ErrInvalidArgument)
		}
		hashedPassword, err := passwords.HashPassword(*input.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		admin.Password = hashedPassword
	}

	if err := s.adminRepo.Update(ctx, admin); err != nil {
		return nil, fmt.Errorf("failed to update admin: %w", err)
	}

	if input.RoleIDs != nil {
		roles, err := s.parseRoles(ctx, input.RoleIDs)
		if err != nil {
			return nil, err
		}
		if err := s.adminRepo.ReplaceRoles(ctx, admin, roles); err != nil {
			return nil, fmt.Errorf("failed to update admin roles: %w", err)
		}
		admin.Roles = roles
	}

	if input.PermissionIDs != nil {
		perms, err := s.parsePermissions(ctx, input.PermissionIDs)
		if err != nil {
			return nil, err
		}
		if err := s.adminRepo.ReplacePermissions(ctx, admin, perms); err != nil {
			return nil, fmt.Errorf("failed to update admin permissions: %w", err)
		}
		admin.Permissions = perms
	}

	return admin, nil
}

func (s *adminService) DeleteAdmin(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("%w: invalid admin ID", apperrors.ErrInvalidArgument)
	}

	admin, err := s.adminRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find admin: %w", err)
	}
	if admin == nil {
		return fmt.Errorf("%w: admin not found", apperrors.ErrNotFound)
	}

	if err := s.adminRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete admin: %w", err)
	}

	return nil
}

func (s *adminService) ListAdmins(ctx context.Context, page, pageSize int, search string) ([]models.Admin, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.adminRepo.List(ctx, page, pageSize, strings.TrimSpace(search))
}

func (s *adminService) AssignRoles(ctx context.Context, adminID uuid.UUID, roleIDs []string) (*models.Admin, error) {
	if adminID == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid admin ID", apperrors.ErrInvalidArgument)
	}

	admin, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return nil, fmt.Errorf("failed to find admin: %w", err)
	}
	if admin == nil {
		return nil, fmt.Errorf("%w: admin not found", apperrors.ErrNotFound)
	}

	roles, err := s.parseRoles(ctx, roleIDs)
	if err != nil {
		return nil, err
	}

	if err := s.adminRepo.ReplaceRoles(ctx, admin, roles); err != nil {
		return nil, fmt.Errorf("failed to replace roles: %w", err)
	}

	admin.Roles = roles
	return admin, nil
}

func (s *adminService) AssignPermissions(ctx context.Context, adminID uuid.UUID, permissionIDs []string) (*models.Admin, error) {
	if adminID == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid admin ID", apperrors.ErrInvalidArgument)
	}

	admin, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return nil, fmt.Errorf("failed to find admin: %w", err)
	}
	if admin == nil {
		return nil, fmt.Errorf("%w: admin not found", apperrors.ErrNotFound)
	}

	perms, err := s.parsePermissions(ctx, permissionIDs)
	if err != nil {
		return nil, err
	}

	if err := s.adminRepo.ReplacePermissions(ctx, admin, perms); err != nil {
		return nil, fmt.Errorf("failed to replace permissions: %w", err)
	}

	admin.Permissions = perms
	return admin, nil
}
