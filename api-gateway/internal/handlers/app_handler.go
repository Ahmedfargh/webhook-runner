package handlers

import (
	"net/http"
	"strconv"

	pb "webhookApiGateway/api/proto/runner"
	"webhookApiGateway/internal/clients"
	"webhookApiGateway/internal/middleware"

	"github.com/gin-gonic/gin"
)

type AppHandler struct {
	runnerClient *clients.RunnerClient
}

func NewAppHandler(runnerClient *clients.RunnerClient) *AppHandler {
	return &AppHandler{runnerClient: runnerClient}
}

type CreateAppInput struct {
	UserID     string `json:"user_id"`
	Name       string `json:"name" binding:"required"`
	WebhookURL string `json:"webhook_url"`
}

func (h *AppHandler) CreateApp(c *gin.Context) {
	userID := c.GetString("user_id")
	userRole := c.GetString("user_role")
	var input CreateAppInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload: " + err.Error()})
		return
	}

	targetUserID := userID
	if (userRole == "admin" || userRole == "administrator") && input.UserID != "" {
		targetUserID = input.UserID
	}

	res, err := h.runnerClient.App.CreateApp(c.Request.Context(), &pb.CreateAppRequest{
		UserId:     targetUserID,
		Name:       input.Name,
		WebhookUrl: input.WebhookURL,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": res.App})
}

func (h *AppHandler) GetApp(c *gin.Context) {
	userID := c.GetString("user_id")
	role := c.GetString("user_role")
	if role == "admin" || role == "administrator" {
		userID = ""
	}
	id := c.Param("id")

	res, err := h.runnerClient.App.GetApp(c.Request.Context(), &pb.GetAppRequest{
		Id:     id,
		UserId: userID,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": res.App})
}

func (h *AppHandler) ListApps(c *gin.Context) {
	userID := c.GetString("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	role := c.GetString("user_role")

	filterUserID := userID
	if role == "admin" || role == "administrator" {
		if targetUser := c.Query("user_id"); targetUser != "" {
			filterUserID = targetUser
		} else {
			filterUserID = ""
		}
	}

	res, err := h.runnerClient.App.ListApps(c.Request.Context(), &pb.ListAppsRequest{
		UserId: filterUserID,
		Page:   int32(page),
		Limit:  int32(limit),
		Search: search,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  res.Apps,
		"total": res.Total,
		"page":  res.Page,
		"limit": res.Limit,
	})
}

type UpdateAppInput struct {
	Name       string `json:"name"`
	WebhookURL string `json:"webhook_url"`
	IsActive   *bool  `json:"is_active"`
}

func (h *AppHandler) UpdateApp(c *gin.Context) {
	userID := c.GetString("user_id")
	role := c.GetString("user_role")
	if role == "admin" || role == "administrator" {
		userID = ""
	}
	id := c.Param("id")

	var input UpdateAppInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload: " + err.Error()})
		return
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	res, err := h.runnerClient.App.UpdateApp(c.Request.Context(), &pb.UpdateAppRequest{
		Id:         id,
		UserId:     userID,
		Name:       input.Name,
		WebhookUrl: input.WebhookURL,
		IsActive:   isActive,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": res.App})
}

func (h *AppHandler) DeleteApp(c *gin.Context) {
	userID := c.GetString("user_id")
	role := c.GetString("user_role")
	if role == "admin" || role == "administrator" {
		userID = ""
	}
	id := c.Param("id")

	res, err := h.runnerClient.App.DeleteApp(c.Request.Context(), &pb.DeleteAppRequest{
		Id:     id,
		UserId: userID,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": res.Success, "message": res.Message})
}

type RotateSecretInput struct {
	RotateAppSecret     bool `json:"rotate_app_secret"`
	RotateWebhookSecret bool `json:"rotate_webhook_secret"`
}

func (h *AppHandler) RotateSecrets(c *gin.Context) {
	userID := c.GetString("user_id")
	role := c.GetString("user_role")
	if role == "admin" || role == "administrator" {
		userID = ""
	}
	id := c.Param("id")

	var input RotateSecretInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload: " + err.Error()})
		return
	}

	res, err := h.runnerClient.App.RotateSecrets(c.Request.Context(), &pb.RotateSecretsRequest{
		Id:                  id,
		UserId:              userID,
		RotateAppSecret:     input.RotateAppSecret,
		RotateWebhookSecret: input.RotateWebhookSecret,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": res.App})
}
