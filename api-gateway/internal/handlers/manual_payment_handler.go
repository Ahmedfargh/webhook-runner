package handlers

import (
	"net/http"
	"strconv"

	subscriptionsv1 "webhookApiGateway/api/proto/subscriptions/v1"
	"webhookApiGateway/internal/clients"
	"webhookApiGateway/internal/middleware"

	"github.com/gin-gonic/gin"
)

type ManualPaymentHandler struct {
	client *clients.SubscriptionsClient
}

func NewManualPaymentHandler(client *clients.SubscriptionsClient) *ManualPaymentHandler {
	return &ManualPaymentHandler{client: client}
}

type submitPaymentDTO struct {
	InvoiceID            string  `json:"invoice_id" binding:"required"`
	Amount               float64 `json:"amount"`
	Currency             string  `json:"currency"`
	PaymentMethod        string  `json:"payment_method"`
	TransactionReference string  `json:"transaction_reference" binding:"required"`
	PayerName            string  `json:"payer_name"`
	PayerNotes           string  `json:"payer_notes"`
}

func (h *ManualPaymentHandler) SubmitManualPayment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req submitPaymentDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.client.ManualPayment.SubmitManualPayment(c.Request.Context(), &subscriptionsv1.SubmitManualPaymentRequest{
		InvoiceId:            req.InvoiceID,
		UserId:               userID.(string),
		Amount:               req.Amount,
		Currency:             req.Currency,
		PaymentMethod:        req.PaymentMethod,
		TransactionReference: req.TransactionReference,
		PayerName:            req.PayerName,
		PayerNotes:           req.PayerNotes,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Manual offline payment proof submitted for verification",
		"data":    res,
	})
}

type reviewPaymentDTO struct {
	Approve    bool   `json:"approve"`
	AdminNotes string `json:"admin_notes"`
}

func (h *ManualPaymentHandler) ReviewManualPayment(c *gin.Context) {
	paymentID := c.Param("id")
	adminID, _ := c.Get("admin_id")

	var req reviewPaymentDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminIDStr := ""
	if adminID != nil {
		adminIDStr = adminID.(string)
	}

	res, err := h.client.ManualPayment.ReviewManualPayment(c.Request.Context(), &subscriptionsv1.ReviewManualPaymentRequest{
		PaymentId:  paymentID,
		Approve:    req.Approve,
		AdminNotes: req.AdminNotes,
		AdminId:    adminIDStr,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": res.GetMessage(),
		"data":    res.GetPayment(),
	})
}

func (h *ManualPaymentHandler) ListManualPayments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	search := c.Query("search")

	res, err := h.client.ManualPayment.ListManualPayments(c.Request.Context(), &subscriptionsv1.ListManualPaymentsRequest{
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
		"data": res.GetPayments(),
		"pagination": gin.H{
			"total_items":  res.GetTotalItems(),
			"current_page": res.GetCurrentPage(),
			"total_pages":  res.GetTotalPages(),
		},
	})
}
