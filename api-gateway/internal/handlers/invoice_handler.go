package handlers

import (
	"net/http"
	"strconv"

	subscriptionsv1 "webhookApiGateway/api/proto/subscriptions/v1"
	"webhookApiGateway/internal/clients"
	"webhookApiGateway/internal/middleware"

	"github.com/gin-gonic/gin"
)

type InvoiceHandler struct {
	client *clients.SubscriptionsClient
}

func NewInvoiceHandler(client *clients.SubscriptionsClient) *InvoiceHandler {
	return &InvoiceHandler{client: client}
}

func (h *InvoiceHandler) GetMyInvoices(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	search := c.Query("search")

	res, err := h.client.Invoice.ListInvoices(c.Request.Context(), &subscriptionsv1.ListInvoicesRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		UserId:   userID.(string),
		Status:   status,
		Search:   search,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": res.GetInvoices(),
		"pagination": gin.H{
			"total_items":  res.GetTotalItems(),
			"current_page": res.GetCurrentPage(),
			"total_pages":  res.GetTotalPages(),
		},
	})
}

func (h *InvoiceHandler) ListAllInvoices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	userID := c.Query("user_id")
	status := c.Query("status")
	search := c.Query("search")

	res, err := h.client.Invoice.ListInvoices(c.Request.Context(), &subscriptionsv1.ListInvoicesRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		UserId:   userID,
		Status:   status,
		Search:   search,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": res.GetInvoices(),
		"pagination": gin.H{
			"total_items":  res.GetTotalItems(),
			"current_page": res.GetCurrentPage(),
			"total_pages":  res.GetTotalPages(),
		},
	})
}

func (h *InvoiceHandler) GetInvoice(c *gin.Context) {
	id := c.Param("id")
	res, err := h.client.Invoice.GetInvoice(c.Request.Context(), &subscriptionsv1.GetInvoiceRequest{
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

type manualInvoiceItemDTO struct {
	Description string  `json:"description" binding:"required"`
	Quantity    int32   `json:"quantity" binding:"required"`
	UnitPrice   float64 `json:"unit_price" binding:"required"`
}

type createManualInvoiceDTO struct {
	UserID                  string                 `json:"user_id" binding:"required"`
	SubscriptionID          string                 `json:"subscription_id"`
	Amount                  float64                `json:"amount"`
	Tax                     float64                `json:"tax"`
	Currency                string                 `json:"currency"`
	DueDate                 string                 `json:"due_date"`
	Notes                   string                 `json:"notes"`
	BankAccountInstructions string                 `json:"bank_account_instructions"`
	Items                   []manualInvoiceItemDTO `json:"items"`
}

func (h *InvoiceHandler) CreateManualInvoice(c *gin.Context) {
	var req createManualInvoiceDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items := make([]*subscriptionsv1.CreateInvoiceItemInput, 0, len(req.Items))
	for _, itm := range req.Items {
		items = append(items, &subscriptionsv1.CreateInvoiceItemInput{
			Description: itm.Description,
			Quantity:    itm.Quantity,
			UnitPrice:   itm.UnitPrice,
		})
	}

	res, err := h.client.Invoice.CreateManualInvoice(c.Request.Context(), &subscriptionsv1.CreateManualInvoiceRequest{
		UserId:                  req.UserID,
		SubscriptionId:          req.SubscriptionID,
		Amount:                  req.Amount,
		Tax:                     req.Tax,
		Currency:                req.Currency,
		DueDate:                 req.DueDate,
		Notes:                   req.Notes,
		BankAccountInstructions: req.BankAccountInstructions,
		Items:                   items,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Invoice created successfully",
		"data":    res,
	})
}

type markPaidDTO struct {
	PaymentReference string `json:"payment_reference"`
	PaymentMethod    string `json:"payment_method"`
	AdminNotes       string `json:"admin_notes"`
}

func (h *InvoiceHandler) MarkInvoicePaid(c *gin.Context) {
	id := c.Param("id")
	adminID, _ := c.Get("admin_id")

	var req markPaidDTO
	_ = c.ShouldBindJSON(&req)

	adminIDStr := ""
	if adminID != nil {
		adminIDStr = adminID.(string)
	}

	res, err := h.client.Invoice.MarkInvoicePaid(c.Request.Context(), &subscriptionsv1.MarkInvoicePaidRequest{
		InvoiceId:        id,
		PaymentReference: req.PaymentReference,
		PaymentMethod:    req.PaymentMethod,
		AdminNotes:       req.AdminNotes,
		AdminId:          adminIDStr,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invoice marked as paid successfully",
		"data":    res,
	})
}

type voidInvoiceDTO struct {
	Reason string `json:"reason"`
}

func (h *InvoiceHandler) VoidInvoice(c *gin.Context) {
	id := c.Param("id")
	var req voidInvoiceDTO
	_ = c.ShouldBindJSON(&req)

	res, err := h.client.Invoice.VoidInvoice(c.Request.Context(), &subscriptionsv1.VoidInvoiceRequest{
		InvoiceId: id,
		Reason:    req.Reason,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": res.GetMessage(),
		"data":    res.GetInvoice(),
	})
}
