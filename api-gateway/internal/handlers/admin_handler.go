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

type AdminHandler struct {
	client *clients.AccountsClient
}

func NewAdminHandler(client *clients.AccountsClient) *AdminHandler {
	return &AdminHandler{client: client}
}

type CreateAdminDTO struct {
	Name          string   `json:"name" binding:"required"`
	Email         string   `json:"email" binding:"required,email"`
	Phone         string   `json:"phone" binding:"required"`
	Password      string   `json:"password" binding:"required,min=6"`
	CountryID     string   `json:"country_id"`
	RoleIDs       []string `json:"role_ids"`
	PermissionIDs []string `json:"permission_ids"`
}

type UpdateAdminDTO struct {
	Name          string   `json:"name"`
	Email         string   `json:"email"`
	Phone         string   `json:"phone"`
	Password      *string  `json:"password,omitempty"`
	CountryID     string   `json:"country_id"`
	RoleIDs       []string `json:"role_ids"`
	PermissionIDs []string `json:"permission_ids"`
}

type AssignRolesDTO struct {
	RoleIDs []string `json:"role_ids" binding:"required"`
}

type AssignPermissionsDTO struct {
	PermissionIDs []string `json:"permission_ids" binding:"required"`
}

// ListAdmins retrieves admins with pagination and search
func (h *AdminHandler) ListAdmins(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.Admin.ListAdmins(ctx, &accountsv1.ListAdminsRequest{
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
		"data":       res.Admins,
		"pagination": res.Pagination,
	})
}

// GetAdmin retrieves an admin by ID
func (h *AdminHandler) GetAdmin(c *gin.Context) {
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.Admin.GetAdmin(ctx, &accountsv1.GetAdminRequest{Id: id})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    res,
	})
}

// CreateAdmin creates a new admin
func (h *AdminHandler) CreateAdmin(c *gin.Context) {
	var dto CreateAdminDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.Admin.CreateAdmin(ctx, &accountsv1.CreateAdminRequest{
		Name:          dto.Name,
		Email:         dto.Email,
		Phone:         dto.Phone,
		Password:      dto.Password,
		CountryId:     dto.CountryID,
		RoleIds:       dto.RoleIDs,
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

// UpdateAdmin updates an existing admin
func (h *AdminHandler) UpdateAdmin(c *gin.Context) {
	id := c.Param("id")
	var dto UpdateAdminDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &accountsv1.UpdateAdminRequest{
		Id:            id,
		Name:          dto.Name,
		Email:         dto.Email,
		Phone:         dto.Phone,
		CountryId:     dto.CountryID,
		RoleIds:       dto.RoleIDs,
		PermissionIds: dto.PermissionIDs,
	}
	if dto.Password != nil && *dto.Password != "" {
		req.Password = dto.Password
	}

	res, err := h.client.Admin.UpdateAdmin(ctx, req)
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    res,
	})
}

// DeleteAdmin removes an admin
func (h *AdminHandler) DeleteAdmin(c *gin.Context) {
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.Admin.DeleteAdmin(ctx, &accountsv1.DeleteAdminRequest{Id: id})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": res.Success,
		"message": res.Message,
	})
}

// AssignRoles assigns roles to an admin
func (h *AdminHandler) AssignRoles(c *gin.Context) {
	id := c.Param("id")
	var dto AssignRolesDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.Admin.AssignRolesToAdmin(ctx, &accountsv1.AssignRolesToAdminRequest{
		AdminId: id,
		RoleIds: dto.RoleIDs,
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

// AssignPermissions assigns direct permissions to an admin
func (h *AdminHandler) AssignPermissions(c *gin.Context) {
	id := c.Param("id")
	var dto AssignPermissionsDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.Admin.AssignPermissionsToAdmin(ctx, &accountsv1.AssignPermissionsToAdminRequest{
		AdminId:       id,
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
