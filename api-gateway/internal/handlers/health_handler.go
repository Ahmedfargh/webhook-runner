package handlers

import (
	"net/http"
	"time"

	"webhookApiGateway/internal/clients"
	"webhookApiGateway/internal/config"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	accountsClient      *clients.AccountsClient
	subscriptionsClient *clients.SubscriptionsClient
	runnerClient        *clients.RunnerClient
	cfg                 *config.Config
}

func NewHealthHandler(
	accountsClient *clients.AccountsClient,
	subscriptionsClient *clients.SubscriptionsClient,
	runnerClient *clients.RunnerClient,
	cfg *config.Config,
) *HealthHandler {
	return &HealthHandler{
		accountsClient:      accountsClient,
		subscriptionsClient: subscriptionsClient,
		runnerClient:        runnerClient,
		cfg:                 cfg,
	}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	accLatency, accErr := h.accountsClient.Ping(c.Request.Context())
	accountsStatus := "healthy"
	var accountsErrMsg string
	if accErr != nil {
		accountsStatus = "degraded"
		accountsErrMsg = accErr.Error()
	}

	subLatency, subErr := h.subscriptionsClient.Ping(c.Request.Context())
	subscriptionsStatus := "healthy"
	var subErrMsg string
	if subErr != nil {
		subscriptionsStatus = "degraded"
		subErrMsg = subErr.Error()
	}

	runLatency, runErr := h.runnerClient.Ping(c.Request.Context())
	runnerStatus := "healthy"
	var runErrMsg string
	if runErr != nil {
		runnerStatus = "degraded"
		runErrMsg = runErr.Error()
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
				"latency_ms":       accLatency.Milliseconds(),
				"error":            accountsErrMsg,
			},
			"subscriptions_grpc": gin.H{
				"status":           subscriptionsStatus,
				"host":             h.cfg.SubscriptionsGRPCHost,
				"port":             h.cfg.SubscriptionsGRPCPort,
				"service_identity": h.cfg.ServiceName,
				"latency_ms":       subLatency.Milliseconds(),
				"error":            subErrMsg,
			},
			"runner_grpc": gin.H{
				"status":           runnerStatus,
				"host":             h.cfg.RunnerGRPCHost,
				"port":             h.cfg.RunnerGRPCPort,
				"service_identity": h.cfg.ServiceName,
				"latency_ms":       runLatency.Milliseconds(),
				"error":            runErrMsg,
			},
		},
	}

	c.JSON(http.StatusOK, response)
}
