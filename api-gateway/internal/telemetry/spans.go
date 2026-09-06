package telemetry

import (
	"context"
	"sync"
	"time"
)

type contextKey string

const spanCollectorKey contextKey = "request_span_collector"

type Span struct {
	Name       string  `json:"name"`
	Service    string  `json:"service"`
	Protocol   string  `json:"protocol"` // REST, GRPC, KAFKA
	Type       string  `json:"type"`     // INGRESS, DOWNSTREAM_RPC, EVENT_STREAM, EGRESS
	OffsetMs   float64 `json:"offset_ms"`
	DurationMs float64 `json:"duration_ms"`
	Status     string  `json:"status"` // OK, ERROR
	Details    string  `json:"details,omitempty"`
}

type SpanCollector struct {
	mu        sync.Mutex
	startTime time.Time
	traceID   string
	spans     []Span
}

func NewSpanCollector(traceID string) *SpanCollector {
	return &SpanCollector{
		startTime: time.Now(),
		traceID:   traceID,
		spans:     make([]Span, 0, 8),
	}
}

func (c *SpanCollector) GetTraceID() string {
	if c == nil {
		return ""
	}
	return c.traceID
}

func (c *SpanCollector) AddSpan(name, service, protocol, spanType string, start time.Time, duration time.Duration, status, details string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	offset := float64(start.Sub(c.startTime).Microseconds()) / 1000.0
	if offset < 0 {
		offset = 0
	}
	durMs := float64(duration.Microseconds()) / 1000.0
	if durMs < 0.05 {
		durMs = 0.05
	}

	c.spans = append(c.spans, Span{
		Name:       name,
		Service:    service,
		Protocol:   protocol,
		Type:       spanType,
		OffsetMs:   offset,
		DurationMs: durMs,
		Status:     status,
		Details:    details,
	})
}

func (c *SpanCollector) GetSpans() []Span {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	res := make([]Span, len(c.spans))
	copy(res, c.spans)
	return res
}

func WithSpanCollector(ctx context.Context, collector *SpanCollector) context.Context {
	return context.WithValue(ctx, spanCollectorKey, collector)
}

func GetSpanCollector(ctx context.Context) *SpanCollector {
	if ctx == nil {
		return nil
	}
	if val, ok := ctx.Value(spanCollectorKey).(*SpanCollector); ok {
		return val
	}
	return nil
}
