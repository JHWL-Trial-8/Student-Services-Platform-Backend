package adminuserapi

import (
	"net/http"
	"strconv"

	adminusersvc "student-services-platform-backend/app/services/adminuser"
	"student-services-platform-backend/internal/openapi"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *adminusersvc.Service
}

func New(s *adminusersvc.Service) *Handler {
	return &Handler{svc: s}
}

// ListUsers handles GET /users - List users with pagination and role filtering
func (h *Handler) ListUsers(c *gin.Context) {
	// Parse query parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, openapi.Error{
			Code:    "bad_request",
			Message: "Invalid page parameter",
		})
		return
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		c.JSON(http.StatusBadRequest, openapi.Error{
			Code:    "bad_request",
			Message: "Invalid page_size parameter (must be between 1 and 100)",
		})
		return
	}

	// Parse role filter if provided
	var role *openapi.Role
	if roleStr := c.Query("role"); roleStr != "" {
		r := openapi.Role(roleStr)
		// Validate role value
		if r != openapi.STUDENT && r != openapi.ADMIN && r != openapi.SUPER_ADMIN {
			c.JSON(http.StatusBadRequest, openapi.Error{
				Code:    "bad_request",
				Message: "Invalid role parameter",
			})
			return
		}
		role = &r
	}

	// Get users
	result, err := h.svc.ListUsers(page, pageSize, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, openapi.Error{
			Code:    "internal_error",
			Message: "Failed to list users",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetUser handles GET /users/{id} - Get a user by ID
func (h *Handler) GetUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, openapi.Error{
			Code:    "bad_request",
			Message: "User ID is required",
		})
		return
	}

	user, err := h.svc.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, openapi.Error{
			Code:    "not_found",
			Message: "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// CreateUser handles POST /users - Create a new user
func (h *Handler) CreateUser(c *gin.Context) {
	var req openapi.UserCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, openapi.Error{
			Code:    "bad_request",
			Message: "Invalid request body",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Validate required fields
	if req.Email == "" || req.Name == "" || req.Role == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, openapi.Error{
			Code:    "bad_request",
			Message: "Email, name, role, and password are required",
		})
		return
	}

	// Validate role value
	if req.Role != openapi.STUDENT && req.Role != openapi.ADMIN && req.Role != openapi.SUPER_ADMIN {
		c.JSON(http.StatusBadRequest, openapi.Error{
			Code:    "bad_request",
			Message: "Invalid role value",
		})
		return
	}

	user, err := h.svc.CreateUser(req)
	if err != nil {
		switch e := err.(type) {
		case *adminusersvc.ErrEmailTaken:
			c.JSON(http.StatusBadRequest, openapi.Error{
				Code:    "email_taken",
				Message: "Email is already taken",
				Details: map[string]interface{}{"email": e.Email},
			})
		default:
			c.JSON(http.StatusInternalServerError, openapi.Error{
				Code:    "internal_error",
				Message: "Failed to create user",
				Details: map[string]interface{}{"error": err.Error()},
			})
		}
		return
	}

	c.JSON(http.StatusCreated, user)
}

// UpdateUser handles PUT /users/{id} - Update a user
func (h *Handler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, openapi.Error{
			Code:    "bad_request",
			Message: "User ID is required",
		})
		return
	}

	var req openapi.UserAdminUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, openapi.Error{
			Code:    "bad_request",
			Message: "Invalid request body",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Validate required fields
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, openapi.Error{
			Code:    "bad_request",
			Message: "Email is required",
		})
		return
	}

	// Validate role value if provided
	if req.Role != "" && req.Role != openapi.STUDENT && req.Role != openapi.ADMIN && req.Role != openapi.SUPER_ADMIN {
		c.JSON(http.StatusBadRequest, openapi.Error{
			Code:    "bad_request",
			Message: "Invalid role value",
		})
		return
	}

	user, err := h.svc.UpdateUser(id, req)
	if err != nil {
		switch e := err.(type) {
		case *adminusersvc.ErrEmailTaken:
			c.JSON(http.StatusBadRequest, openapi.Error{
				Code:    "email_taken",
				Message: "Email is already taken",
				Details: map[string]interface{}{"email": e.Email},
			})
		default:
			c.JSON(http.StatusInternalServerError, openapi.Error{
				Code:    "internal_error",
				Message: "Failed to update user",
				Details: map[string]interface{}{"error": err.Error()},
			})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser handles DELETE /users/{id} - Delete a user
func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, openapi.Error{
			Code:    "bad_request",
			Message: "User ID is required",
		})
		return
	}

	if err := h.svc.DeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, openapi.Error{
			Code:    "internal_error",
			Message: "Failed to delete user",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	c.Status(http.StatusNoContent)
}