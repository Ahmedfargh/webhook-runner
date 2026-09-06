package middleware

import (
	"bytes"
	"context"
	"io"
	"strings"
	"time"

	"webhookApiGateway/internal/audit"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AuditMiddleware(emitter *audit.KafkaEmitter) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.Request.URL.Path

		// Only audit mutating actions or auth endpoints, skip health/metrics/reading audit logs
		isMutating := method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE"
		if !isMutating || strings.HasPrefix(path, "/api/v1/audit-logs") || path == "/health" || path == "/api/v1/health" {
			c.Next()
			return
		}

		// Read request body copy
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		start := time.Now()
		c.Next()

		// Post-processing: extract actor & outcome
		actorID := c.GetString("user_id")
		actorEmail := c.GetString("user_email")
		actorRole := c.GetString("user_role")

		actorType := "USER"
		if actorRole == "admin" {
			actorType = "ADMIN"
		} else if actorID == "" {
			actorType = "ANONYMOUS"
		}

		// Determine Resource and Action from path
		resource, action := resolveResourceAndAction(method, path)

		status := "SUCCESS"
		var errorMsg string
		if c.Writer.Status() >= 400 {
			status = "FAILED"
			if len(c.Errors) > 0 {
				errorMsg = c.Errors.String()
			}
		}

		// Clean payload: mask passwords if present
		cleanPayload := sanitizePayload(string(bodyBytes))

		event := audit.Event{
			ID:           uuid.New().String(),
			ActorID:      actorID,
			ActorType:    actorType,
			ActorName:    c.GetString("user_name"),
			ActorEmail:   actorEmail,
			ServiceName:  "api-gateway",
			Action:       action,
			Resource:     resource,
			ResourceID:   c.Param("id"),
			AfterJSON:    cleanPayload,
			IPAddress:    c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Status:       status,
			ErrorMessage: errorMsg,
			Timestamp:    start.Unix(),
		}

		if emitter != nil {
			emitter.Emit(context.Background(), event)
		}
	}
}

func resolveResourceAndAction(method, path string) (string, string) {
	cleanPath := strings.TrimPrefix(path, "/api/v1/")
	parts := strings.Split(cleanPath, "/")
	rootResource := "SYSTEM"
	if len(parts) > 0 && parts[0] != "" {
		rootResource = strings.ToUpper(parts[0])
	}

	action := "EXECUTE"
	switch method {
	case "POST":
		action = "CREATE"
		if strings.Contains(path, "login") {
			action = "LOGIN"
		} else if strings.Contains(path, "rotate-secrets") {
			action = "ROTATE_SECRET"
		} else if strings.Contains(path, "override") {
			action = "OVERRIDE"
		} else if strings.Contains(path, "dispatch") || strings.Contains(path, "send") {
			action = "DISPATCH"
		} else if strings.Contains(path, "retry") {
			action = "RETRY"
		}
	case "PUT", "PATCH":
		action = "UPDATE"
	case "DELETE":
		action = "DELETE"
	}

	return rootResource, action
}

func sanitizePayload(payload string) string {
	if strings.Contains(payload, "password") {
		// basic masking to avoid storing raw passwords in audit logs
		return `{"payload": "[MASKED_SECURITY_FIELDS]"}`
	}
	return payload
}
