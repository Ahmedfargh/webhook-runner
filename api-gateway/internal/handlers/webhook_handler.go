package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	pb "webhookApiGateway/api/proto/runner"
	"webhookApiGateway/internal/clients"
	"webhookApiGateway/internal/middleware"

	"github.com/gin-gonic/gin"
)

type WebhookHandler struct {
	runnerClient *clients.RunnerClient
}

func NewWebhookHandler(runnerClient *clients.RunnerClient) *WebhookHandler {
	return &WebhookHandler{runnerClient: runnerClient}
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
	AppID         string            `json:"app_id" binding:"required"`
	EventName     string            `json:"event_name" binding:"required"`
	Payload       interface{}       `json:"payload"`
	CustomHeaders map[string]string `json:"custom_headers"`
	TargetURL     string            `json:"target_url_override"`
}

func (h *WebhookHandler) SendWebhook(c *gin.Context) {
	var input SendWebhookInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload: " + err.Error()})
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
	if role == "admin" && c.Query("all") == "true" {
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
	id := c.Param("id")

	res, err := h.runnerClient.Webhook.RetryWebhookCall(c.Request.Context(), &pb.RetryWebhookCallRequest{
		Id:     id,
		UserId: userID,
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
