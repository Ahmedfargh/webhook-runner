package repository

import (
	"context"
	"errors"

	"accounts/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CountryRepository provides data access operations for countries
type CountryRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Country, error)
	FindByCode(ctx context.Context, code string) (*models.Country, error)
	List(ctx context.Context) ([]models.Country, error)
}

type countryRepository struct {
	db *gorm.DB
}

// NewCountryRepository creates a new CountryRepository
func NewCountryRepository(db *gorm.DB) CountryRepository {
	return &countryRepository{db: db}
}

func (r *countryRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Country, error) {
	var country models.Country
	if err := r.db.WithContext(ctx).First(&country, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &country, nil
}

func (r *countryRepository) FindByCode(ctx context.Context, code string) (*models.Country, error) {
	var country models.Country
	if err := r.db.WithContext(ctx).First(&country, "country_code = ?", code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &country, nil
}

func (r *countryRepository) List(ctx context.Context) ([]models.Country, error) {
	var countries []models.Country
	err := r.db.WithContext(ctx).Order("country_code ASC").Find(&countries).Error
	return countries, err
}
