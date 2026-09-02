package repository

import (
	"context"
	"errors"

	"accounts/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleRepository defines data access operations for roles
type RoleRepository interface {
	Create(ctx context.Context, role *models.Role) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Role, error)
	FindByName(ctx context.Context, name string) (*models.Role, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Role, error)
	Update(ctx context.Context, role *models.Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, pageSize int, search string) ([]models.Role, int64, error)
	ReplacePermissions(ctx context.Context, role *models.Role, permissions []models.Permission) error
}

type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new GORM-backed RoleRepository
func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	var role models.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&role, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&role, "name = ?", name).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Role, error) {
	var roles []models.Role
	if len(ids) == 0 {
		return roles, nil
	}
	err := r.db.WithContext(ctx).Preload("Permissions").Where("id IN ?", ids).Find(&roles).Error
	return roles, err
}

func (r *roleRepository) Update(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Select("Permissions").Delete(&models.Role{}, "id = ?", id).Error
}

func (r *roleRepository) List(ctx context.Context, page, pageSize int, search string) ([]models.Role, int64, error) {
	var roles []models.Role
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Role{})
	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Permissions").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

func (r *roleRepository) ReplacePermissions(ctx context.Context, role *models.Role, permissions []models.Permission) error {
	return r.db.WithContext(ctx).Model(role).Association("Permissions").Replace(permissions)
}
