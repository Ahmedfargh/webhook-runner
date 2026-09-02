package service

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"accounts/internal/helpers/passwords"
	"accounts/internal/helpers/phonenumbers"
	"accounts/internal/models"
	"accounts/internal/modules/user/repository"
	"accounts/internal/pkg/apperrors"
	repo "accounts/internal/repository"

	"github.com/google/uuid"
)

// CreateUserInput carries payload for creating a new user
type CreateUserInput struct {
	Name      string
	Email     string
	Phone     string
	Password  string
	CountryID string
}

// UpdateUserInput carries payload for updating an existing user
type UpdateUserInput struct {
	ID        uuid.UUID
	Name      string
	Email     string
	Phone     string
	Password  *string
	CountryID string
}

// UserService defines business logic for users
type UserService interface {
	CreateUser(ctx context.Context, input CreateUserInput) (*models.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	UpdateUser(ctx context.Context, input UpdateUserInput) (*models.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, page, pageSize int, search string) ([]models.User, int64, error)
}

type userService struct {
	userRepo    repository.UserRepository
	countryRepo repo.CountryRepository
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo repository.UserRepository, countryRepo repo.CountryRepository) UserService {
	return &userService{
		userRepo:    userRepo,
		countryRepo: countryRepo,
	}
}

func (s *userService) validateEmail(email string) (string, error) {
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

func (s *userService) CreateUser(ctx context.Context, input CreateUserInput) (*models.User, error) {
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

	// Check if email already registered
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, fmt.Errorf("%w: email '%s' is already taken", apperrors.ErrEmailAlreadyUsed, email)
	}

	hashedPassword, err := passwords.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		ID:        uuid.New(),
		Name:      name,
		Email:     email,
		Phone:     normalizedPhone,
		Password:  hashedPassword,
		CountryID: countryID,
		Country:   *country,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *userService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid user ID", apperrors.ErrInvalidArgument)
	}

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("%w: user not found", apperrors.ErrNotFound)
	}

	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, input UpdateUserInput) (*models.User, error) {
	if input.ID == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid user ID", apperrors.ErrInvalidArgument)
	}

	user, err := s.userRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("%w: user not found", apperrors.ErrNotFound)
	}

	if input.Name != "" {
		user.Name = strings.TrimSpace(input.Name)
	}

	if input.Email != "" {
		email, err := s.validateEmail(input.Email)
		if err != nil {
			return nil, err
		}

		existingUser, err := s.userRepo.FindByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != user.ID {
			return nil, fmt.Errorf("%w: email '%s' is already taken", apperrors.ErrEmailAlreadyUsed, email)
		}
		user.Email = email
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
		user.CountryID = countryID
		user.Country = *country
	}

	if input.Phone != "" {
		normalizedPhone, err := phonenumbers.NormalizePhoneNumber(input.Phone, user.Country.CountryCode)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", apperrors.ErrPhoneInvalid, err)
		}
		user.Phone = normalizedPhone
	}

	if input.Password != nil && *input.Password != "" {
		if len(*input.Password) < 6 {
			return nil, fmt.Errorf("%w: password must be at least 6 characters", apperrors.ErrInvalidArgument)
		}
		hashedPassword, err := passwords.HashPassword(*input.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = hashedPassword
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("%w: invalid user ID", apperrors.ErrInvalidArgument)
	}

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("%w: user not found", apperrors.ErrNotFound)
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (s *userService) ListUsers(ctx context.Context, page, pageSize int, search string) ([]models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.userRepo.List(ctx, page, pageSize, strings.TrimSpace(search))
}
