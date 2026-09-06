package handlers

import (
	"net/http"
	"strconv"

	pb "webhookApiGateway/api/proto/audit/v1"
	"webhookApiGateway/internal/clients"
	"webhookApiGateway/internal/middleware"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	auditClient *clients.AuditClient
}

func NewAuditHandler(auditClient *clients.AuditClient) *AuditHandler {
	return &AuditHandler{auditClient: auditClient}
}

func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	actorID := c.Query("actor_id")
	serviceName := c.Query("service_name")
	resource := c.Query("resource")
	action := c.Query("action")
	status := c.Query("status")
	search := c.Query("search")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	res, err := h.auditClient.Audit.ListAuditLogs(c.Request.Context(), &pb.ListAuditLogsRequest{
		ActorId:     actorID,
		ServiceName: serviceName,
		Resource:    resource,
		Action:      action,
		Status:      status,
		Search:      search,
		StartDate:   startDate,
		EndDate:     endDate,
		Page:        int32(page),
		Limit:       int32(limit),
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  res.AuditLogs,
		"total": res.Total,
		"page":  res.Page,
		"limit": res.Limit,
	})
}

func (h *AuditHandler) GetAuditLog(c *gin.Context) {
	id := c.Param("id")

	res, err := h.auditClient.Audit.GetAuditLog(c.Request.Context(), &pb.GetAuditLogRequest{
		Id: id,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": res.AuditLog,
	})
}
