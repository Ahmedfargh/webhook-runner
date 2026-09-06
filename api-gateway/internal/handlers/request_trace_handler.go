package handlers

import (
	"net/http"
	"strconv"

	pb "webhookApiGateway/api/proto/request_tracker/v1"
	"webhookApiGateway/internal/clients"
	"webhookApiGateway/internal/middleware"

	"github.com/gin-gonic/gin"
)

type RequestTraceHandler struct {
	trackerClient *clients.RequestTrackerClient
}

func NewRequestTraceHandler(trackerClient *clients.RequestTrackerClient) *RequestTraceHandler {
	return &RequestTraceHandler{trackerClient: trackerClient}
}

func (h *RequestTraceHandler) ListTraces(c *gin.Context) {
	if h.trackerClient.Tracker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Request tracker service unavailable"})
		return
	}

	actorType := c.Query("actor_type")
	actorID := c.Query("actor_id")
	method := c.Query("method")
	route := c.Query("route")
	search := c.Query("search")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	statusCode, _ := strconv.Atoi(c.Query("status_code"))
	minLifetime, _ := strconv.ParseFloat(c.Query("min_lifetime_ms"), 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	res, err := h.trackerClient.Tracker.ListTraces(c.Request.Context(), &pb.ListTracesRequest{
		ActorType:     actorType,
		ActorId:       actorID,
		Method:        method,
		Route:         route,
		StatusCode:    int32(statusCode),
		MinLifetimeMs: minLifetime,
		Search:        search,
		StartDate:     startDate,
		EndDate:       endDate,
		Page:          int32(page),
		Limit:         int32(limit),
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  res.Traces,
		"total": res.Total,
		"page":  res.Page,
		"limit": res.Limit,
	})
}

func (h *RequestTraceHandler) GetTrace(c *gin.Context) {
	if h.trackerClient.Tracker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Request tracker service unavailable"})
		return
	}

	id := c.Param("id")

	res, err := h.trackerClient.Tracker.GetTrace(c.Request.Context(), &pb.GetTraceRequest{
		Id: id,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": res.Trace,
	})
}

func (h *RequestTraceHandler) GetStats(c *gin.Context) {
	if h.trackerClient.Tracker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Request tracker service unavailable"})
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	res, err := h.trackerClient.Tracker.GetTraceStats(c.Request.Context(), &pb.GetTraceStatsRequest{
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"total_requests":    res.TotalRequests,
			"avg_lifetime_ms":   res.AvgLifetimeMs,
			"p95_lifetime_ms":   res.P95LifetimeMs,
			"p99_lifetime_ms":   res.P99LifetimeMs,
			"error_count":       res.ErrorCount,
			"error_rate":        res.ErrorRate,
		},
	})
}
