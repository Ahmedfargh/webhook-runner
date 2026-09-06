package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	pb "webhookApiGateway/api/proto/runner"
	"webhookApiGateway/internal/clients"
	"webhookApiGateway/internal/kafka"
	"webhookApiGateway/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WebhookHandler struct {
	runnerClient  *clients.RunnerClient
	kafkaProducer *kafka.KafkaProducer
}

func NewWebhookHandler(runnerClient *clients.RunnerClient, kafkaProducer *kafka.KafkaProducer) *WebhookHandler {
	return &WebhookHandler{
		runnerClient:  runnerClient,
		kafkaProducer: kafkaProducer,
	}
}

// TestReceiver is a built-in mock destination webhook endpoint for local development & testing
func (h *WebhookHandler) TestReceiver(c *gin.Context) {
	var body interface{}
	_ = c.ShouldBindJSON(&body)

	sig := c.GetHeader("X-Webhook-Signature")
	event := c.GetHeader("X-Webhook-Event")
	webhookID := c.GetHeader("X-Webhook-ID")
	ts := c.GetHeader("X-Webhook-Timestamp")

	c.JSON(http.StatusOK, gin.H{
		"status":      "received",
		"message":     "Webhook received and verified successfully by local test receiver",
		"received_at": time.Now().Format(time.RFC3339),
		"webhook_id":  webhookID,
		"event":       event,
		"timestamp":   ts,
		"signature":   sig,
		"payload":     body,
	})
}

type SendWebhookInput struct {
	AppID         string            `json:"app_id"`
	EventName     string            `json:"event_name"`
	Payload       interface{}       `json:"payload"`
	CustomHeaders map[string]string `json:"custom_headers"`
	TargetURL     string            `json:"target_url_override"`
	Async         *bool             `json:"async,omitempty"`
}

func (h *WebhookHandler) SendWebhook(c *gin.Context) {
	var input SendWebhookInput
	_ = c.ShouldBindJSON(&input)

	// Fallback to URL query parameters if not present in request body
	if input.AppID == "" {
		input.AppID = c.Query("app_id")
		if input.AppID == "" {
			input.AppID = c.Query("appId")
		}
	}
	if input.EventName == "" {
		input.EventName = c.Query("event_name")
		if input.EventName == "" {
			input.EventName = c.Query("event")
		}
	}
	if input.TargetURL == "" {
		input.TargetURL = c.Query("target_url_override")
		if input.TargetURL == "" {
			input.TargetURL = c.Query("target_url")
			if input.TargetURL == "" {
				input.TargetURL = c.Query("url")
			}
		}
	}
	if input.Payload == nil && c.Query("payload") != "" {
		input.Payload = c.Query("payload")
	}

	if input.AppID == "" || input.EventName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app_id and event_name are required (in JSON body or query parameters)"})
		return
	}

	payloadJSON := "{}"
	if input.Payload != nil {
		if str, ok := input.Payload.(string); ok {
			payloadJSON = str
		} else {
			bytes, _ := json.Marshal(input.Payload)
			payloadJSON = string(bytes)
		}
	}

	// Determine dispatch mode: Kafka vs Synchronous gRPC
	forceSync := c.Query("sync") == "true" || c.Query("sync") == "1" || (input.Async != nil && !*input.Async)
	useKafka := h.kafkaProducer != nil && h.kafkaProducer.IsEnabled() && !forceSync

	if useKafka {
		callID := fmt.Sprintf("wh_call_%s", strings.ReplaceAll(uuid.New().String(), "-", "")[:16])
		event := &kafka.WebhookDispatchEvent{
			CallID:            callID,
			AppID:             input.AppID,
			EventName:         input.EventName,
			PayloadJSON:       payloadJSON,
			CustomHeaders:     input.CustomHeaders,
			TargetURLOverride: input.TargetURL,
			CreatedAt:         time.Now().UTC(),
			Timestamp:         time.Now().Unix(),
		}

		err := h.kafkaProducer.PublishWebhookDispatch(c.Request.Context(), event)
		if err == nil {
			c.JSON(http.StatusAccepted, gin.H{
				"data": gin.H{
					"id":                  callID,
					"app_id":              input.AppID,
					"event_name":          input.EventName,
					"status":              "QUEUED",
					"target_url_override": input.TargetURL,
					"dispatch_mode":       "kafka_stream",
					"timestamp":           event.Timestamp,
				},
				"success": true,
				"message": "Webhook dispatch event published to Kafka queue for processing",
			})
			return
		}

		// Graceful fallback to gRPC on Kafka write error
		log.Printf("[API Gateway] Kafka publish failed (%v), falling back to synchronous gRPC dispatch\n", err)
	}

	// Synchronous gRPC dispatch (Direct or Fallback)
	res, err := h.runnerClient.Webhook.SendWebhook(c.Request.Context(), &pb.SendWebhookRequest{
		AppId:             input.AppID,
		EventName:         input.EventName,
		PayloadJson:       payloadJSON,
		CustomHeaders:     input.CustomHeaders,
		TargetUrlOverride: input.TargetURL,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    res.WebhookCall,
		"success": res.Success,
		"message": res.Message,
	})
}

func (h *WebhookHandler) ListWebhookCalls(c *gin.Context) {
	userID := c.GetString("user_id")
	appID := c.Query("app_id")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")

	role := c.GetString("user_role")
	filterUserID := userID
	if role == "admin" || appID != "" {
		filterUserID = ""
	}

	res, err := h.runnerClient.Webhook.ListWebhookCalls(c.Request.Context(), &pb.ListWebhookCallsRequest{
		UserId: filterUserID,
		AppId:  appID,
		Status: status,
		Page:   int32(page),
		Limit:  int32(limit),
		Search: search,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  res.WebhookCalls,
		"total": res.Total,
		"page":  res.Page,
		"limit": res.Limit,
	})
}

func (h *WebhookHandler) GetWebhookCall(c *gin.Context) {
	userID := c.GetString("user_id")
	role := c.GetString("user_role")
	if role == "admin" || role == "administrator" {
		userID = ""
	}
	id := c.Param("id")

	res, err := h.runnerClient.Webhook.GetWebhookCall(c.Request.Context(), &pb.GetWebhookCallRequest{
		Id:     id,
		UserId: userID,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": res.WebhookCall})
}

func (h *WebhookHandler) RetryWebhookCall(c *gin.Context) {
	userID := c.GetString("user_id")
	role := c.GetString("user_role")
	if role == "admin" || role == "administrator" {
		userID = ""
	}
	id := c.Param("id")

	res, err := h.runnerClient.Webhook.RetryWebhookCall(c.Request.Context(), &pb.RetryWebhookCallRequest{
		Id:     id,
		UserId: userID,
	})
	if err != nil {
		fmt.Println(err)
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    res.WebhookCall,
		"success": res.Success,
		"message": res.Message,
	})
}
