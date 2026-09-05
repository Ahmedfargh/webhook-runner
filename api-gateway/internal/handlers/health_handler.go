package handlers

import (
	"net/http"
	"time"

	"webhookApiGateway/internal/clients"
	"webhookApiGateway/internal/config"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	client *clients.AccountsClient
	cfg    *config.Config
}

func NewHealthHandler(client *clients.AccountsClient, cfg *config.Config) *HealthHandler {
	return &HealthHandler{client: client, cfg: cfg}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	latency, err := h.client.Ping(c.Request.Context())

	accountsStatus := "healthy"
	var accountsErr string
	if err != nil {
		accountsStatus = "degraded"
		accountsErr = err.Error()
	}

	response := gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "webhook-api-gateway",
		"upstream_services": gin.H{
			"accounts_grpc": gin.H{
				"status":           accountsStatus,
				"host":             h.cfg.AccountsGRPCHost,
				"port":             h.cfg.AccountsGRPCPort,
				"service_identity": h.cfg.ServiceName,
				"latency_ms":       latency.Milliseconds(),
				"error":            accountsErr,
			},
		},
	}

	c.JSON(http.StatusOK, response)
}
