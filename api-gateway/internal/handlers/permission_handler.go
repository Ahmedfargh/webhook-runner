package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	accountsv1 "webhookApiGateway/api/proto/v1"
	"webhookApiGateway/internal/clients"
	"webhookApiGateway/internal/middleware"

	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	client *clients.AccountsClient
}

func NewPermissionHandler(client *clients.AccountsClient) *PermissionHandler {
	return &PermissionHandler{client: client}
}

type CreatePermissionDTO struct {
	Name string `json:"name" binding:"required"`
}

type UpdatePermissionDTO struct {
	Name string `json:"name" binding:"required"`
}

// ListPermissions retrieves all permissions
func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	search := c.Query("search")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.Permission.ListPermissions(ctx, &accountsv1.ListPermissionsRequest{
		Pagination: &accountsv1.PaginationRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
			Search:   search,
		},
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       res.Permissions,
		"pagination": res.Pagination,
	})
}

// GetPermission retrieves a single permission
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.Permission.GetPermission(ctx, &accountsv1.GetPermissionRequest{Id: id})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    res,
	})
}

// CreatePermission creates a new permission
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var dto CreatePermissionDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.Permission.CreatePermission(ctx, &accountsv1.CreatePermissionRequest{
		Name: dto.Name,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    res,
	})
}

// UpdatePermission updates an existing permission
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	id := c.Param("id")
	var dto UpdatePermissionDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.Permission.UpdatePermission(ctx, &accountsv1.UpdatePermissionRequest{
		Id:   id,
		Name: dto.Name,
	})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    res,
	})
}

// DeletePermission removes a permission
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.Permission.DeletePermission(ctx, &accountsv1.DeletePermissionRequest{Id: id})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": res.Success,
		"message": res.Message,
	})
}
