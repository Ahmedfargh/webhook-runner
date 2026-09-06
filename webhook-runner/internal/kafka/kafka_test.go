package kafka

import (
	"context"
	"testing"
	"time"

	"webhookRunner/internal/models"
)

type mockProcessor struct {
	called bool
	event  *WebhookDispatchEvent
}

func (m *mockProcessor) ProcessAsyncDispatch(ctx context.Context, event *WebhookDispatchEvent) (*models.WebhookCall, error) {
	m.called = true
	m.event = event
	return &models.WebhookCall{
		EventName:          event.EventName,
		Status:             models.StatusSuccess,
		ResponseStatusCode: 200,
	}, nil
}

func TestConsumerDisabled(t *testing.T) {
	mock := &mockProcessor{}
	consumer := NewConsumer("localhost:9092", "webhook-dispatches", "test-group", false, mock, nil)
	if consumer.IsEnabled() {
		t.Fatalf("expected consumer to be disabled")
	}

	err := consumer.Ping(context.Background())
	if err == nil {
		t.Fatalf("expected ping error on disabled consumer")
	}
}

func TestResultProducerDisabled(t *testing.T) {
	producer := NewResultProducer("localhost:9092", "webhook-results", false)
	if producer.IsEnabled() {
		t.Fatalf("expected result producer to be disabled")
	}

	err := producer.PublishResult(context.Background(), &WebhookResultEvent{
		CallID: "test_call",
		Status: "SUCCESS",
	})
	if err != nil {
		t.Fatalf("expected nil error on disabled result producer, got: %v", err)
	}
}

func TestEventsModel(t *testing.T) {
	res := WebhookResultEvent{
		CallID:             "wh_123",
		AppID:              "app_456",
		EventName:          "payment.success",
		TargetURL:          "https://example.com/receiver",
		Status:             "SUCCESS",
		ResponseStatusCode: 200,
		ResponseLatencyMS:  45,
		AttemptCount:       1,
		CompletedAt:        time.Now().UTC(),
	}

	if res.Status != "SUCCESS" || res.ResponseStatusCode != 200 {
		t.Errorf("expected success status and 200 code")
	}
}
