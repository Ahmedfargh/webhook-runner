package kafka

import "time"

// WebhookDispatchEvent represents a webhook dispatch request message received from Kafka
type WebhookDispatchEvent struct {
	CallID            string            `json:"call_id"`
	AppID             string            `json:"app_id"`
	EventName         string            `json:"event_name"`
	PayloadJSON       string            `json:"payload_json"`
	CustomHeaders     map[string]string `json:"custom_headers,omitempty"`
	TargetURLOverride string            `json:"target_url_override,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	Timestamp         int64             `json:"timestamp"`
}

// WebhookResultEvent represents the result of a dispatched webhook published by Webhook Runner
type WebhookResultEvent struct {
	CallID             string    `json:"call_id"`
	AppID              string    `json:"app_id"`
	EventName          string    `json:"event_name"`
	TargetURL          string    `json:"target_url"`
	Status             string    `json:"status"` // SUCCESS, FAILED
	ResponseStatusCode int32     `json:"response_status_code"`
	ResponseBody       string    `json:"response_body,omitempty"`
	ResponseLatencyMS  int64     `json:"response_latency_ms"`
	ErrorMessage       string    `json:"error_message,omitempty"`
	AttemptCount       int32     `json:"attempt_count"`
	CompletedAt        time.Time `json:"completed_at"`
}
