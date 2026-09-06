package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"webhookApiGateway/internal/telemetry"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type traceBodyWriter struct {
	gin.ResponseWriter
	body    *bytes.Buffer
	maxSize int
}

func (w *traceBodyWriter) Write(b []byte) (int, error) {
	if w.body.Len() < w.maxSize {
		remaining := w.maxSize - w.body.Len()
		if len(b) > remaining {
			w.body.Write(b[:remaining])
			w.body.WriteString("... [truncated]")
		} else {
			w.body.Write(b)
		}
	}
	return w.ResponseWriter.Write(b)
}

func (w *traceBodyWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func RequestTrackerMiddleware(emitter *telemetry.RequestTrackerEmitter) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Optionally skip high-frequency internal health checks if desired, but capture all API traffic
		if path == "/health" || path == "/api/v1/health" {
			c.Next()
			return
		}

		start := time.Now()

		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = c.GetHeader("X-Trace-ID")
		}
		if reqID == "" {
			reqID = "req-" + uuid.New().String()
		}

		c.Header("X-Request-ID", reqID)
		c.Header("X-Trace-ID", reqID)
		c.Set("request_id", reqID)
		c.Set("trace_id", reqID)

		// Initialize trip span collector and attach to request context
		collector := telemetry.NewSpanCollector(reqID)
		collector.AddSpan("REST Ingress & Gateway Routing", "api-gateway", "REST", "INGRESS", start, 250*time.Microsecond, "OK", "Method: "+c.Request.Method)
		c.Request = c.Request.WithContext(telemetry.WithSpanCollector(c.Request.Context(), collector))

		// 1. Read and buffer Request Body
		var reqBodyBytes []byte
		if c.Request.Body != nil {
			reqBodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
		}

		// 2. Wrap Response Writer to intercept status code and response payload
		bodyBuffer := bytes.NewBuffer(make([]byte, 0, 1024))
		trackedWriter := &traceBodyWriter{
			ResponseWriter: c.Writer,
			body:           bodyBuffer,
			maxSize:        8192, // 8KB cap
		}
		c.Writer = trackedWriter

		// Execute downstream handlers & microservices
		c.Next()

		completedAt := time.Now()
		lifetimeMs := float64(completedAt.Sub(start).Microseconds()) / 1000.0

		// Add final Gateway Egress span
		egressDuration := time.Since(completedAt)
		if egressDuration < 150*time.Microsecond {
			egressDuration = 150 * time.Microsecond
		}
		collector.AddSpan("REST Response Egress", "api-gateway", "REST", "EGRESS", completedAt, egressDuration, "OK", "")

		// Serialize the complete trip waterfall
		var spansJSON string
		if spans := collector.GetSpans(); len(spans) > 0 {
			if b, err := json.Marshal(spans); err == nil {
				spansJSON = string(b)
			}
		}

		// 3. Extract identity from context
		actorID := c.GetString("user_id")
		actorEmail := c.GetString("user_email")
		actorRole := c.GetString("user_role")
		actorName := c.GetString("user_name")

		actorType := "ANONYMOUS"
		if actorRole == "admin" || actorRole == "super_admin" {
			actorType = "ADMIN"
		} else if actorID != "" {
			actorType = "USER"
		}

		// 4. Sanitize Request and Response Bodies
		sanitizedReq := sanitizeTracePayload(string(reqBodyBytes))
		sanitizedResp := sanitizeTracePayload(bodyBuffer.String())

		var errMsg string
		if len(c.Errors) > 0 {
			errMsg = c.Errors.String()
		}

		route := c.FullPath()
		if route == "" {
			route = path
		}

		// 5. Emit trace asynchronously with full trip breakdown
		if emitter != nil {
			emitter.EmitAsync(telemetry.TracePayload{
				ID:           uuid.New().String(),
				TraceID:      reqID,
				RequestID:    reqID,
				ActorType:    actorType,
				ActorID:      actorID,
				ActorName:    actorName,
				ActorEmail:   actorEmail,
				ActorRole:    actorRole,
				ServiceName:  "api-gateway",
				Method:       c.Request.Method,
				Path:         path,
				Route:        route,
				QueryParams:  c.Request.URL.RawQuery,
				ClientIP:     c.ClientIP(),
				UserAgent:    c.Request.UserAgent(),
				StatusCode:   c.Writer.Status(),
				LifetimeMs:   lifetimeMs,
				RequestBody:  sanitizedReq,
				ResponseBody: sanitizedResp,
				ErrorMessage: errMsg,
				SpansJSON:    spansJSON,
				ReceivedAt:   start.UTC().Format(time.RFC3339Nano),
				CompletedAt:  completedAt.UTC().Format(time.RFC3339Nano),
			})
		}
	}
}

// sanitizeTracePayload removes sensitive credentials, passwords, tokens from recorded payloads
func sanitizeTracePayload(payload string) string {
	if payload == "" {
		return ""
	}

	sensitiveKeys := []string{"password", "token", "secret", "authorization", "refresh_token", "credit_card"}
	lowered := strings.ToLower(payload)
	for _, key := range sensitiveKeys {
		if strings.Contains(lowered, key) {
			var obj map[string]interface{}
			if err := json.Unmarshal([]byte(payload), &obj); err == nil {
				maskMap(obj, sensitiveKeys)
				if maskedBytes, err := json.Marshal(obj); err == nil {
					return string(maskedBytes)
				}
			}
			return `{"masked": "[SENSITIVE_CREDENTIALS_HIDDEN]"}`
		}
	}

	return payload
}

func maskMap(data map[string]interface{}, sensitiveKeys []string) {
	for k, v := range data {
		kLower := strings.ToLower(k)
		isSensitive := false
		for _, s := range sensitiveKeys {
			if strings.Contains(kLower, s) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			data[k] = "******"
		} else if childMap, ok := v.(map[string]interface{}); ok {
			maskMap(childMap, sensitiveKeys)
		}
	}
}
