package engine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"webhookRunner/internal/helpers"
)

type DispatchRequest struct {
	CallID        string
	EventName     string
	TargetURL     string
	PayloadJSON   string
	WebhookSecret string
	CustomHeaders map[string]string
	Timeout       time.Duration
}

type DispatchResult struct {
	StatusCode int32
	Body       string
	LatencyMS  int64
	Success    bool
	Error      string
	Headers    map[string]string
}

type Dispatcher struct {
	client *http.Client
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        500,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (d *Dispatcher) Dispatch(ctx context.Context, req DispatchRequest) DispatchResult {
	startTime := time.Now()

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodPost, req.TargetURL, bytes.NewBuffer([]byte(req.PayloadJSON)))
	if err != nil {
		latency := time.Since(startTime).Milliseconds()
		return DispatchResult{
			StatusCode: 0,
			LatencyMS:  latency,
			Success:    false,
			Error:      fmt.Sprintf("Invalid target request: %v", err),
		}
	}

	// Calculate HMAC signature
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signedPayload := []byte(fmt.Sprintf("%s.%s", timestamp, req.PayloadJSON))
	signature := helpers.ComputeHMACSHA256(signedPayload, req.WebhookSecret)
	sigHeader := fmt.Sprintf("t=%s,v1=%s", timestamp, signature)

	// Set standard headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "Webhook-Runner/1.0 (+https://webhook.io)")
	httpReq.Header.Set("X-Webhook-ID", req.CallID)
	httpReq.Header.Set("X-Webhook-Event", req.EventName)
	httpReq.Header.Set("X-Webhook-Timestamp", timestamp)
	httpReq.Header.Set("X-Webhook-Signature", sigHeader)

	for k, v := range req.CustomHeaders {
		httpReq.Header.Set(k, v)
	}

	resp, err := d.client.Do(httpReq)
	latency := time.Since(startTime).Milliseconds()

	if err != nil {
		return DispatchResult{
			StatusCode: 0,
			LatencyMS:  latency,
			Success:    false,
			Error:      fmt.Sprintf("HTTP dispatch failed: %v", err),
		}
	}
	defer resp.Body.Close()

	// Read response snippet (limit to 4KB)
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	bodyStr := string(bodyBytes)

	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	return DispatchResult{
		StatusCode: int32(resp.StatusCode),
		Body:       bodyStr,
		LatencyMS:  latency,
		Success:    success,
		Error:      "",
	}
}
