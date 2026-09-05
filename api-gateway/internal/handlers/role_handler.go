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

type RoleHandler struct {
	client *clients.AccountsClient
}

func NewRoleHandler(client *clients.AccountsClient) *RoleHandler {
	return &RoleHandler{client: client}
}

type CreateRoleDTO struct {
	Name          string   `json:"name" binding:"required"`
	PermissionIDs []string `json:"permission_ids"`
}

type UpdateRoleDTO struct {
	Name          string   `json:"name" binding:"required"`
	PermissionIDs []string `json:"permission_ids"`
}

type AssignRolePermissionsDTO struct {
	PermissionIDs []string `json:"permission_ids" binding:"required"`
}

// ListRoles retrieves all roles
func (h *RoleHandler) ListRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.Role.ListRoles(ctx, &accountsv1.ListRolesRequest{
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
		"data":       res.Roles,
		"pagination": res.Pagination,
	})
}

// GetRole retrieves a single role
func (h *RoleHandler) GetRole(c *gin.Context) {
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.Role.GetRole(ctx, &accountsv1.GetRoleRequest{Id: id})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    res,
	})
}

// CreateRole creates a new role
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var dto CreateRoleDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.Role.CreateRole(ctx, &accountsv1.CreateRoleRequest{
		Name:          dto.Name,
		PermissionIds: dto.PermissionIDs,
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

// UpdateRole updates an existing role
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	var dto UpdateRoleDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.Role.UpdateRole(ctx, &accountsv1.UpdateRoleRequest{
		Id:            id,
		Name:          dto.Name,
		PermissionIds: dto.PermissionIDs,
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

// DeleteRole deletes a role
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.Role.DeleteRole(ctx, &accountsv1.DeleteRoleRequest{Id: id})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": res.Success,
		"message": res.Message,
	})
}

// AssignPermissions assigns permissions to a role
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	id := c.Param("id")
	var dto AssignRolePermissionsDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.Role.AssignPermissionsToRole(ctx, &accountsv1.AssignPermissionsToRoleRequest{
		RoleId:        id,
		PermissionIds: dto.PermissionIDs,
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
