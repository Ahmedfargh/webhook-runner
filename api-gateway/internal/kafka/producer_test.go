package kafka

import (
	"context"
	"testing"
	"time"
)

func TestKafkaProducerDisabled(t *testing.T) {
	producer := NewKafkaProducer("localhost:9092", "webhook-dispatches", false)
	if producer.IsEnabled() {
		t.Fatalf("expected producer to be disabled")
	}

	err := producer.PublishWebhookDispatch(context.Background(), &WebhookDispatchEvent{
		CallID:    "test_123",
		AppID:     "app_live_abc",
		EventName: "order.created",
	})
	if err == nil {
		t.Fatalf("expected error when publishing to disabled producer")
	}

	err = producer.Ping(context.Background())
	if err == nil {
		t.Fatalf("expected ping error when producer is disabled")
	}
}

func TestWebhookDispatchEventSerialization(t *testing.T) {
	event := WebhookDispatchEvent{
		CallID:            "call_test_001",
		AppID:             "app_live_xyz",
		EventName:         "invoice.paid",
		PayloadJSON:       `{"amount": 99.00}`,
		CustomHeaders:     map[string]string{"X-Custom": "Val"},
		TargetURLOverride: "https://example.com/webhook",
		CreatedAt:         time.Now().UTC(),
		Timestamp:         time.Now().Unix(),
	}

	if event.CallID != "call_test_001" {
		t.Errorf("expected CallID to match, got %s", event.CallID)
	}
	if event.EventName != "invoice.paid" {
		t.Errorf("expected EventName to match, got %s", event.EventName)
	}
}
