package handlers

import (
	"net/http"

	subscriptionsv1 "webhookApiGateway/api/proto/subscriptions/v1"
	"webhookApiGateway/internal/clients"
	"webhookApiGateway/internal/middleware"

	"github.com/gin-gonic/gin"
)

type PlanHandler struct {
	client *clients.SubscriptionsClient
}

func NewPlanHandler(client *clients.SubscriptionsClient) *PlanHandler {
	return &PlanHandler{client: client}
}

func (h *PlanHandler) ListPlans(c *gin.Context) {
	includeInactive := c.Query("include_inactive") == "true"
	res, err := h.client.Plan.ListPlans(c.Request.Context(), &subscriptionsv1.ListPlansRequest{
		IncludeInactive: includeInactive,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": res.GetPlans(),
	})
}

func (h *PlanHandler) GetPlan(c *gin.Context) {
	id := c.Param("id")
	res, err := h.client.Plan.GetPlan(c.Request.Context(), &subscriptionsv1.GetPlanRequest{
		Id: id,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": res,
	})
}

type createPlanDTO struct {
	Name              string   `json:"name" binding:"required"`
	Code              string   `json:"code"`
	Description       string   `json:"description"`
	PriceMonthly      float64  `json:"price_monthly"`
	PriceAnnually     float64  `json:"price_annually"`
	Currency          string   `json:"currency"`
	MaxWebhooks       int32    `json:"max_webhooks"`
	MaxEventsPerMonth int64    `json:"max_events_per_month"`
	MaxTeamMembers    int32    `json:"max_team_members"`
	Features          []string `json:"features"`
	IsActive          bool     `json:"is_active"`
	IsPopular         bool     `json:"is_popular"`
	TierLevel         int32    `json:"tier_level"`
}

func (h *PlanHandler) CreatePlan(c *gin.Context) {
	var req createPlanDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.client.Plan.CreatePlan(c.Request.Context(), &subscriptionsv1.CreatePlanRequest{
		Name:              req.Name,
		Code:              req.Code,
		Description:       req.Description,
		PriceMonthly:      req.PriceMonthly,
		PriceAnnually:     req.PriceAnnually,
		Currency:          req.Currency,
		MaxWebhooks:       req.MaxWebhooks,
		MaxEventsPerMonth: req.MaxEventsPerMonth,
		MaxTeamMembers:    req.MaxTeamMembers,
		Features:          req.Features,
		IsActive:          req.IsActive,
		IsPopular:         req.IsPopular,
		TierLevel:         req.TierLevel,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Plan created successfully",
		"data":    res,
	})
}

func (h *PlanHandler) UpdatePlan(c *gin.Context) {
	id := c.Param("id")
	var req createPlanDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.client.Plan.UpdatePlan(c.Request.Context(), &subscriptionsv1.UpdatePlanRequest{
		Id:                id,
		Name:              req.Name,
		Description:       req.Description,
		PriceMonthly:      req.PriceMonthly,
		PriceAnnually:     req.PriceAnnually,
		Currency:          req.Currency,
		MaxWebhooks:       req.MaxWebhooks,
		MaxEventsPerMonth: req.MaxEventsPerMonth,
		MaxTeamMembers:    req.MaxTeamMembers,
		Features:          req.Features,
		IsActive:          req.IsActive,
		IsPopular:         req.IsPopular,
		TierLevel:         req.TierLevel,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Plan updated successfully",
		"data":    res,
	})
}

func (h *PlanHandler) DeletePlan(c *gin.Context) {
	id := c.Param("id")
	res, err := h.client.Plan.DeletePlan(c.Request.Context(), &subscriptionsv1.DeletePlanRequest{
		Id: id,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": res.GetMessage(),
	})
}
