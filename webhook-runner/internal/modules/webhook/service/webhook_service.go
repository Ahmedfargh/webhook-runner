package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"webhookRunner/internal/engine"
	"webhookRunner/internal/models"
	appRepo "webhookRunner/internal/modules/app/repository"
	"webhookRunner/internal/modules/webhook/repository"

	"github.com/google/uuid"
)

type WebhookService interface {
	SendWebhook(ctx context.Context, appIDStr, eventName, payloadJSON string, customHeaders map[string]string, targetURLOverride string) (*models.WebhookCall, error)
	ListWebhookCalls(ctx context.Context, userID uuid.UUID, appIDStr string, status string, page, limit int, search string) ([]models.WebhookCall, int64, error)
	GetWebhookCall(ctx context.Context, id, userID uuid.UUID) (*models.WebhookCall, error)
	RetryWebhookCall(ctx context.Context, id, userID uuid.UUID) (*models.WebhookCall, error)
}

type webhookService struct {
	repo       repository.WebhookRepository
	appRepo    appRepo.AppRepository
	dispatcher *engine.Dispatcher
}

func NewWebhookService(repo repository.WebhookRepository, aRepo appRepo.AppRepository, dispatcher *engine.Dispatcher) WebhookService {
	return &webhookService{
		repo:       repo,
		appRepo:    aRepo,
		dispatcher: dispatcher,
	}
}

func (s *webhookService) SendWebhook(ctx context.Context, appIDStr, eventName, payloadJSON string, customHeaders map[string]string, targetURLOverride string) (*models.WebhookCall, error) {
	if appIDStr == "" {
		return nil, errors.New("app_id is required")
	}
	if eventName == "" {
		return nil, errors.New("event_name is required")
	}
	if payloadJSON == "" {
		payloadJSON = "{}"
	}

	// 1. Resolve App by either UUID or AppID string
	var app *models.App
	var err error
	if parsedUUID, parseErr := uuid.Parse(appIDStr); parseErr == nil {
		app, err = s.appRepo.FindByID(parsedUUID)
	} else {
		app, err = s.appRepo.FindByAppID(appIDStr)
	}

	if err != nil || app == nil {
		return nil, fmt.Errorf("app not found with identifier %s: %v", appIDStr, err)
	}

	if !app.IsActive {
		return nil, errors.New("app is currently deactivated")
	}

	targetURL := app.WebhookURL
	if targetURLOverride != "" {
		targetURL = targetURLOverride
	}

	headersBytes, _ := json.Marshal(customHeaders)

	// 2. Create initial WebhookCall record (Pending)
	callID := uuid.New()
	call := &models.WebhookCall{
		ID:           callID,
		AppID:        app.ID,
		EventName:    eventName,
		TargetURL:    targetURL,
		PayloadJSON:  payloadJSON,
		HeadersJSON:  string(headersBytes),
		AttemptCount: 1,
		Status:       models.StatusPending,
	}

	if err := s.repo.Create(call); err != nil {
		return nil, fmt.Errorf("failed to create webhook call log: %v", err)
	}

	// 3. Dispatch Webhook HTTP POST
	dispatchReq := engine.DispatchRequest{
		CallID:        callID.String(),
		EventName:     eventName,
		TargetURL:     targetURL,
		PayloadJSON:   payloadJSON,
		WebhookSecret: app.WebhookSecret,
		CustomHeaders: customHeaders,
	}

	result := s.dispatcher.Dispatch(ctx, dispatchReq)

	// 4. Update WebhookCall record with response telemetry
	if result.Success {
		call.Status = models.StatusSuccess
	} else {
		call.Status = models.StatusFailed
	}
	call.ResponseStatusCode = result.StatusCode
	call.ResponseBody = result.Body
	call.ResponseLatencyMS = result.LatencyMS
	call.ErrorMessage = result.Error

	if err := s.repo.Update(call); err != nil {
		return nil, fmt.Errorf("failed to update webhook call telemetry: %v", err)
	}

	call.App = *app
	return call, nil
}

func (s *webhookService) ListWebhookCalls(ctx context.Context, userID uuid.UUID, appIDStr string, status string, page, limit int, search string) ([]models.WebhookCall, int64, error) {
	var appID uuid.UUID
	if appIDStr != "" {
		if parsed, err := uuid.Parse(appIDStr); err == nil {
			appID = parsed
		} else if app, err := s.appRepo.FindByAppID(appIDStr); err == nil && app != nil {
			appID = app.ID
		}
	}
	return s.repo.List(userID, appID, status, page, limit, search)
}

func (s *webhookService) GetWebhookCall(ctx context.Context, id, userID uuid.UUID) (*models.WebhookCall, error) {
	call, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if userID != uuid.Nil && call.App.UserID != userID {
		return nil, errors.New("unauthorized: webhook call belongs to another user")
	}
	return call, nil
}

func (s *webhookService) RetryWebhookCall(ctx context.Context, id, userID uuid.UUID) (*models.WebhookCall, error) {
	call, err := s.GetWebhookCall(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	call.AttemptCount++
	call.Status = models.StatusRetrying

	var customHeaders map[string]string
	if call.HeadersJSON != "" {
		_ = json.Unmarshal([]byte(call.HeadersJSON), &customHeaders)
	}

	dispatchReq := engine.DispatchRequest{
		CallID:        call.ID.String(),
		EventName:     call.EventName,
		TargetURL:     call.TargetURL,
		PayloadJSON:   call.PayloadJSON,
		WebhookSecret: call.App.WebhookSecret,
		CustomHeaders: customHeaders,
	}

	result := s.dispatcher.Dispatch(ctx, dispatchReq)

	if result.Success {
		call.Status = models.StatusSuccess
	} else {
		call.Status = models.StatusFailed
	}
	call.ResponseStatusCode = result.StatusCode
	call.ResponseBody = result.Body
	call.ResponseLatencyMS = result.LatencyMS
	call.ErrorMessage = result.Error

	if err := s.repo.Update(call); err != nil {
		return nil, err
	}

	return call, nil
}
