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

type UserHandler struct {
	client *clients.AccountsClient
}

func NewUserHandler(client *clients.AccountsClient) *UserHandler {
	return &UserHandler{client: client}
}

type CreateUserDTO struct {
	Name      string `json:"name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Phone     string `json:"phone" binding:"required"`
	Password  string `json:"password" binding:"required,min=6"`
	CountryID string `json:"country_id"`
}

type UpdateUserDTO struct {
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	Phone     string  `json:"phone"`
	Password  *string `json:"password,omitempty"`
	CountryID string  `json:"country_id"`
}

// ListUsers retrieves users with pagination and search
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.User.ListUsers(ctx, &accountsv1.ListUsersRequest{
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
		"data":       res.Users,
		"pagination": res.Pagination,
	})
}

// GetUser retrieves a user by ID
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.User.GetUser(ctx, &accountsv1.GetUserRequest{Id: id})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    res,
	})
}

// CreateUser creates a new user
func (h *UserHandler) CreateUser(c *gin.Context) {
	var dto CreateUserDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.User.CreateUser(ctx, &accountsv1.CreateUserRequest{
		Name:      dto.Name,
		Email:     dto.Email,
		Phone:     dto.Phone,
		Password:  dto.Password,
		CountryId: dto.CountryID,
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

// UpdateUser updates an existing user
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var dto UpdateUserDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &accountsv1.UpdateUserRequest{
		Id:        id,
		Name:      dto.Name,
		Email:     dto.Email,
		Phone:     dto.Phone,
		CountryId: dto.CountryID,
	}
	if dto.Password != nil && *dto.Password != "" {
		req.Password = dto.Password
	}

	res, err := h.client.User.UpdateUser(ctx, req)
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    res,
	})
}

// DeleteUser removes a user
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.client.User.DeleteUser(ctx, &accountsv1.DeleteUserRequest{Id: id})
	if err != nil {
		middleware.HandleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": res.Success,
		"message": res.Message,
	})
}
