package service

import (
	"context"
	"errors"
	"net/url"

	"webhookRunner/internal/helpers"
	"webhookRunner/internal/models"
	"webhookRunner/internal/modules/app/repository"

	"github.com/google/uuid"
)

type AppService interface {
	CreateApp(ctx context.Context, userID uuid.UUID, name, webhookURL string) (*models.App, error)
	GetApp(ctx context.Context, id uuid.UUID, userID uuid.UUID, appID string) (*models.App, error)
	ListApps(ctx context.Context, userID uuid.UUID, page, limit int, search string) ([]models.App, int64, error)
	UpdateApp(ctx context.Context, id uuid.UUID, userID uuid.UUID, name, webhookURL string, isActive bool) (*models.App, error)
	DeleteApp(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	RotateSecrets(ctx context.Context, id uuid.UUID, userID uuid.UUID, rotateAppSecret, rotateWebhookSecret bool) (*models.App, error)
}

type appService struct {
	repo repository.AppRepository
}

func NewAppService(repo repository.AppRepository) AppService {
	return &appService{repo: repo}
}

func (s *appService) CreateApp(ctx context.Context, userID uuid.UUID, name, webhookURL string) (*models.App, error) {
	if name == "" {
		return nil, errors.New("application name is required")
	}
	if webhookURL == "" {
		return nil, errors.New("webhook destination URL is required")
	}

	// Validate Webhook URL format
	parsedURL, err := url.ParseRequestURI(webhookURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return nil, errors.New("invalid webhook URL (must be valid http/https)")
	}

	app := &models.App{
		UserID:        userID,
		Name:          name,
		AppID:         helpers.GenerateAppID(),
		AppSecret:     helpers.GenerateAppSecret(),
		WebhookURL:    webhookURL,
		WebhookSecret: helpers.GenerateWebhookSecret(),
		IsActive:      true,
	}

	if err := s.repo.Create(app); err != nil {
		return nil, err
	}

	return app, nil
}

func (s *appService) GetApp(ctx context.Context, id uuid.UUID, userID uuid.UUID, appID string) (*models.App, error) {
	if appID != "" {
		return s.repo.FindByAppID(appID)
	}
	if id != uuid.Nil {
		app, err := s.repo.FindByID(id)
		if err != nil {
			return nil, err
		}
		if userID != uuid.Nil && app.UserID != userID {
			return nil, errors.New("unauthorized: app belongs to another user")
		}
		return app, nil
	}
	return nil, errors.New("app id or identifier required")
}

func (s *appService) ListApps(ctx context.Context, userID uuid.UUID, page, limit int, search string) ([]models.App, int64, error) {
	return s.repo.ListByUserID(userID, page, limit, search)
}

func (s *appService) UpdateApp(ctx context.Context, id uuid.UUID, userID uuid.UUID, name, webhookURL string, isActive bool) (*models.App, error) {
	app, err := s.GetApp(ctx, id, userID, "")
	if err != nil {
		return nil, err
	}

	if name != "" {
		app.Name = name
	}
	if webhookURL != "" {
		parsedURL, err := url.ParseRequestURI(webhookURL)
		if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
			return nil, errors.New("invalid webhook URL (must be valid http/https)")
		}
		app.WebhookURL = webhookURL
	}
	app.IsActive = isActive

	if err := s.repo.Update(app); err != nil {
		return nil, err
	}

	return app, nil
}

func (s *appService) DeleteApp(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.Delete(id, userID)
}

func (s *appService) RotateSecrets(ctx context.Context, id uuid.UUID, userID uuid.UUID, rotateAppSecret, rotateWebhookSecret bool) (*models.App, error) {
	app, err := s.GetApp(ctx, id, userID, "")
	if err != nil {
		return nil, err
	}

	if rotateAppSecret {
		app.AppSecret = helpers.GenerateAppSecret()
	}
	if rotateWebhookSecret {
		app.WebhookSecret = helpers.GenerateWebhookSecret()
	}

	if err := s.repo.Update(app); err != nil {
		return nil, err
	}

	return app, nil
}
