package handlers

import (
	"context"
	"net/http"
	"time"

	accountsv1 "webhookApiGateway/api/proto/v1"
	"webhookApiGateway/internal/clients"
	"webhookApiGateway/internal/middleware"

	"github.com/gin-gonic/gin"
)

type CountryHandler struct {
	client *clients.AccountsClient
}

func NewCountryHandler(client *clients.AccountsClient) *CountryHandler {
	return &CountryHandler{client: client}
}

// ListCountries fetches all countries from Accounts gRPC service
func (h *CountryHandler) ListCountries(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	search := c.Query("search")
	res, err := h.client.Country.ListCountries(ctx, &accountsv1.ListCountriesRequest{
		Search: search,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    res.Countries,
	})
}
