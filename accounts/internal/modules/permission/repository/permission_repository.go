package repository

import (
	"context"
	"errors"

	"accounts/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionRepository defines the data access contract for permissions
type PermissionRepository interface {
	Create(ctx context.Context, permission *models.Permission) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Permission, error)
	FindByName(ctx context.Context, name string) (*models.Permission, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Permission, error)
	Update(ctx context.Context, permission *models.Permission) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, pageSize int, search string) ([]models.Permission, int64, error)
}

type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a new GORM-backed PermissionRepository
func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) Create(ctx context.Context, permission *models.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

func (r *permissionRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	var permission models.Permission
	if err := r.db.WithContext(ctx).First(&permission, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) FindByName(ctx context.Context, name string) (*models.Permission, error) {
	var permission models.Permission
	if err := r.db.WithContext(ctx).First(&permission, "name = ?", name).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Permission, error) {
	var permissions []models.Permission
	if len(ids) == 0 {
		return permissions, nil
	}
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) Update(ctx context.Context, permission *models.Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

func (r *permissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Permission{}, "id = ?", id).Error
}

func (r *permissionRepository) List(ctx context.Context, page, pageSize int, search string) ([]models.Permission, int64, error) {
	var permissions []models.Permission
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Permission{})
	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&permissions).Error; err != nil {
		return nil, 0, err
	}

	return permissions, total, nil
}
