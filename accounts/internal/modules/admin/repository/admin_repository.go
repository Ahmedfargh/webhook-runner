package repository

import (
	"context"
	"errors"

	"accounts/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminRepository defines data access operations for admins
type AdminRepository interface {
	Create(ctx context.Context, admin *models.Admin) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Admin, error)
	FindByEmail(ctx context.Context, email string) (*models.Admin, error)
	Update(ctx context.Context, admin *models.Admin) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, pageSize int, search string) ([]models.Admin, int64, error)
	ReplaceRoles(ctx context.Context, admin *models.Admin, roles []models.Role) error
	ReplacePermissions(ctx context.Context, admin *models.Admin, permissions []models.Permission) error
}

type adminRepository struct {
	db *gorm.DB
}

// NewAdminRepository creates a new GORM-backed AdminRepository
func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) Create(ctx context.Context, admin *models.Admin) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

func (r *adminRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Admin, error) {
	var admin models.Admin
	if err := r.db.WithContext(ctx).
		Preload("Country").
		Preload("Roles.Permissions").
		Preload("Permissions").
		First(&admin, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) FindByEmail(ctx context.Context, email string) (*models.Admin, error) {
	var admin models.Admin
	if err := r.db.WithContext(ctx).
		Preload("Country").
		Preload("Roles.Permissions").
		Preload("Permissions").
		First(&admin, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) Update(ctx context.Context, admin *models.Admin) error {
	return r.db.WithContext(ctx).Save(admin).Error
}

func (r *adminRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Select("Roles", "Permissions").Delete(&models.Admin{}, "id = ?", id).Error
}

func (r *adminRepository) List(ctx context.Context, page, pageSize int, search string) ([]models.Admin, int64, error) {
	var admins []models.Admin
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Admin{})
	if search != "" {
		query = query.Where("name LIKE ? OR email LIKE ? OR phone LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.
		Preload("Country").
		Preload("Roles.Permissions").
		Preload("Permissions").
		Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&admins).Error; err != nil {
		return nil, 0, err
	}

	return admins, total, nil
}

func (r *adminRepository) ReplaceRoles(ctx context.Context, admin *models.Admin, roles []models.Role) error {
	return r.db.WithContext(ctx).Model(admin).Association("Roles").Replace(roles)
}

func (r *adminRepository) ReplacePermissions(ctx context.Context, admin *models.Admin, permissions []models.Permission) error {
	return r.db.WithContext(ctx).Model(admin).Association("Permissions").Replace(permissions)
}
