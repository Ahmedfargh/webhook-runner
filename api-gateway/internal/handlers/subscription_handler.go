package handlers

import (
	"net/http"
	"strconv"

	subscriptionsv1 "webhookApiGateway/api/proto/subscriptions/v1"
	"webhookApiGateway/internal/clients"
	"webhookApiGateway/internal/middleware"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	client *clients.SubscriptionsClient
}

func NewSubscriptionHandler(client *clients.SubscriptionsClient) *SubscriptionHandler {
	return &SubscriptionHandler{client: client}
}

func (h *SubscriptionHandler) GetCurrentSubscription(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	res, err := h.client.Subscription.GetUserSubscription(c.Request.Context(), &subscriptionsv1.GetUserSubscriptionRequest{
		UserId: userID.(string),
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": res,
	})
}

type subscribeDTO struct {
	PlanID        string `json:"plan_id" binding:"required"`
	BillingCycle  string `json:"billing_cycle" binding:"required"` // monthly, annually
	PaymentMethod string `json:"payment_method"`
	Notes         string `json:"notes"`
}

func (h *SubscriptionHandler) Subscribe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req subscribeDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.client.Subscription.Subscribe(c.Request.Context(), &subscriptionsv1.SubscribeRequest{
		UserId:        userID.(string),
		PlanId:        req.PlanID,
		BillingCycle:  req.BillingCycle,
		PaymentMethod: req.PaymentMethod,
		Notes:         req.Notes,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription order created successfully",
		"data":    res,
	})
}

type cancelDTO struct {
	Reason      string `json:"reason"`
	Immediately bool   `json:"immediately"`
}

func (h *SubscriptionHandler) CancelSubscription(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req cancelDTO
	_ = c.ShouldBindJSON(&req)

	res, err := h.client.Subscription.CancelSubscription(c.Request.Context(), &subscriptionsv1.CancelSubscriptionRequest{
		UserId:      userID.(string),
		Reason:      req.Reason,
		Immediately: req.Immediately,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": res.GetMessage(),
		"data":    res.GetSubscription(),
	})
}

func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	search := c.Query("search")

	res, err := h.client.Subscription.ListSubscriptions(c.Request.Context(), &subscriptionsv1.ListSubscriptionsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Status:   status,
		Search:   search,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": res.GetSubscriptions(),
		"pagination": gin.H{
			"total_items":  res.GetTotalItems(),
			"current_page": res.GetCurrentPage(),
			"total_pages":  res.GetTotalPages(),
		},
	})
}

type adminOverrideDTO struct {
	UserID           string `json:"user_id" binding:"required"`
	PlanID           string `json:"plan_id"`
	Status           string `json:"status"`
	CurrentPeriodEnd string `json:"current_period_end"`
	AdminNotes       string `json:"admin_notes"`
}

func (h *SubscriptionHandler) AdminOverrideSubscription(c *gin.Context) {
	var req adminOverrideDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.client.Subscription.AdminOverrideSubscription(c.Request.Context(), &subscriptionsv1.AdminOverrideSubscriptionRequest{
		UserId:           req.UserID,
		PlanId:           req.PlanID,
		Status:           req.Status,
		CurrentPeriodEnd: req.CurrentPeriodEnd,
		AdminNotes:       req.AdminNotes,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription override applied successfully",
		"data":    res,
	})
}
