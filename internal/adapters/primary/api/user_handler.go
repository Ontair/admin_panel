package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ontair/admin-panel/internal/core/dto"
	"github.com/ontair/admin-panel/internal/core/entities"
	"github.com/ontair/admin-panel/internal/core/ports/service"
	"go.uber.org/zap"
)

// UserHandler handles user management HTTP requests
type UserHandler struct {
	userService service.UserService
	logger      service.Logger
}

// NewUserHandler creates new user handler
func NewUserHandler(userService service.UserService, logger service.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// RegisterRoutes registers user routes
func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
	users := r.Group("/users")
	{
		// Get current user profile (any authenticated user)
		users.GET("/profile", h.GetCurrentUser)

		// Change password (any authenticated user)
		users.POST("/change-password", h.ChangePassword)
	}
}

// RegisterAdminRoutes registers admin-only user routes
func (h *UserHandler) RegisterAdminRoutes(r *gin.RouterGroup) {
	admin := r.Group("/users")
	{
		// List ALL users (admin only) - полный список со всеми ролями
		admin.GET("/", h.ListAllUsers)

		// Delete user (admin only)
		admin.DELETE("/:id", h.DeleteUser)

		// Activate user (admin only)
		admin.POST("/:id/activate", h.ActivateUser)

		// Deactivate user (admin only)
		admin.POST("/:id/deactivate", h.DeactivateUser)
	}
}

// RegisterManagerRoutes registers manager and admin user routes
func (h *UserHandler) RegisterManagerRoutes(r *gin.RouterGroup) {
	manager := r.Group("/users")
	{
		// List users (manager and admin) - manager sees only user/guest, admin sees all
		manager.GET("/", h.ListUsers)

		// Create user (manager and admin)
		manager.POST("/", h.CreateUser)

		// Get user by ID (manager and admin)
		manager.GET("/:id", h.GetUser)

		// Update user (manager and admin)
		manager.PUT("/:id", h.UpdateUser)
	}
}

// GetCurrentUser returns current user profile
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Internal Server Error",
			"message": "Invalid user ID format",
		})
		return
	}

	// Get user
	user, err := h.userService.GetCurrentUser(c.Request.Context(), userIDUint)
	if err != nil {
		switch err {
		case entities.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Not Found",
				"message": "User not found",
			})
		default:
			h.logger.Error("Get current user failed", zap.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Internal Server Error",
				"message": "Failed to get user",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dto.ToUserDTO(user),
	})
}

// CreateUser creates a new user (admin only)
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.UserCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrBadRequest)
		return
	}

	// Convert DTO to service request
	createReq := &service.CreateUserRequest{
		Username:  req.Username,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      entities.Role(req.Role),
		IsActive:  req.IsActive,
	}

	// Call service
	user, err := h.userService.CreateUser(c.Request.Context(), createReq)
	if err != nil {
		switch err {
		case entities.ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, dto.ErrUserAlreadyExists)
		case entities.ErrInvalidUsername, entities.ErrPasswordTooShort:
			c.JSON(http.StatusBadRequest, dto.ErrValidationFailed)
		default:
			h.logger.Error("Create user failed", zap.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, dto.ErrInternalServer)
		}
		return
	}

	c.JSON(http.StatusCreated, dto.ToUserDTO(user))
}

// GetUser retrieves user by ID
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrBadRequest)
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		switch err {
		case entities.ErrUserNotFound:
			c.JSON(http.StatusNotFound, dto.ErrUserNotFound)
		default:
			h.logger.Error("Get user failed", zap.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, dto.ErrInternalServer)
		}
		return
	}

	c.JSON(http.StatusOK, dto.ToUserDTO(user))
}

// UpdateUser updates user data
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrBadRequest)
		return
	}

	var req dto.UserUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrBadRequest)
		return
	}

	// Convert DTO to service request
	updateReq := &service.UpdateUserRequest{
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      (*entities.Role)(req.Role),
		IsActive:  req.IsActive,
	}

	// Call service
	user, err := h.userService.UpdateUser(c.Request.Context(), uint(id), updateReq)
	if err != nil {
		switch err {
		case entities.ErrUserNotFound:
			c.JSON(http.StatusNotFound, dto.ErrUserNotFound)
		case entities.ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, dto.ErrUserAlreadyExists)
		case entities.ErrInvalidUsername:
			c.JSON(http.StatusBadRequest, dto.ErrValidationFailed)
		default:
			h.logger.Error("Update user failed", zap.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, dto.ErrInternalServer)
		}
		return
	}

	c.JSON(http.StatusOK, dto.ToUserDTO(user))
}

// DeleteUser deletes user by ID (admin only)
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrBadRequest)
		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), uint(id))
	if err != nil {
		switch err {
		case entities.ErrUserNotFound:
			c.JSON(http.StatusNotFound, dto.ErrUserNotFound)
		default:
			h.logger.Error("Delete user failed", zap.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, dto.ErrInternalServer)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ListUsers retrieves paginated list of users (manager view - only user/guest roles)
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	role := c.Query("role")
	search := c.Query("search")
	isActiveStr := c.Query("is_active")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var isActive *bool
	if isActiveStr != "" {
		if isActiveStr == "true" {
			val := true
			isActive = &val
		} else if isActiveStr == "false" {
			val := false
			isActive = &val
		}
	}

	// Manager can only see user and guest roles
	requestedRole := entities.Role(role)
	if requestedRole != "" && requestedRole != entities.RoleUser && requestedRole != entities.RoleGuest {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Forbidden",
			"message": "Manager can only view user and guest roles",
		})
		return
	}

	// Create service request
	listReq := &service.ListUsersRequest{
		Limit:    limit,
		Offset:   offset,
		Role:     requestedRole,
		IsActive: isActive,
		Search:   search,
	}

	// Call service (manager view - only user/guest roles)
	response, err := h.userService.ListUsersForManager(c.Request.Context(), listReq)
	if err != nil {
		h.logger.Error("List users failed", zap.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, dto.ErrInternalServer)
		return
	}

	// Convert to DTOs
	var userDTOs []dto.UserDTO
	for _, user := range response.Users {
		userDTOs = append(userDTOs, dto.ToUserDTO(user))
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  userDTOs,
		"total":  response.Total,
		"limit":  response.Limit,
		"offset": response.Offset,
	})
}

// ListAllUsers retrieves paginated list of ALL users (admin view - all roles)
func (h *UserHandler) ListAllUsers(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	role := c.Query("role")
	search := c.Query("search")
	isActiveStr := c.Query("is_active")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var isActive *bool
	if isActiveStr != "" {
		if isActiveStr == "true" {
			val := true
			isActive = &val
		} else if isActiveStr == "false" {
			val := false
			isActive = &val
		}
	}

	// Create service request (admin can see all roles)
	listReq := &service.ListUsersRequest{
		Limit:    limit,
		Offset:   offset,
		Role:     entities.Role(role),
		IsActive: isActive,
		Search:   search,
	}

	// Call service
	response, err := h.userService.ListUsers(c.Request.Context(), listReq)
	if err != nil {
		h.logger.Error("List all users failed", zap.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, dto.ErrInternalServer)
		return
	}

	// Convert to DTOs
	var userDTOs []dto.UserDTO
	for _, user := range response.Users {
		userDTOs = append(userDTOs, dto.ToUserDTO(user))
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  userDTOs,
		"total":  response.Total,
		"limit":  response.Limit,
		"offset": response.Offset,
	})
}

// ChangePassword allows user to change their password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Internal Server Error",
			"message": "Invalid user ID format",
		})
		return
	}

	var req dto.ChangePasswordDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Bad Request",
			"message": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Convert DTO to service request
	changeReq := &service.ChangePasswordRequest{
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}

	// Call service
	err := h.userService.ChangePassword(c.Request.Context(), userIDUint, changeReq)
	if err != nil {
		switch err {
		case entities.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Not Found",
				"message": "User not found",
			})
		case entities.ErrInvalidCredentials:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Bad Request",
				"message": "Current password is incorrect",
			})
		case entities.ErrPasswordTooShort:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Bad Request",
				"message": "New password is too short",
			})
		default:
			h.logger.Error("Change password failed", zap.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Internal Server Error",
				"message": "Failed to change password",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password changed successfully",
	})
}

// ActivateUser activates user account (admin only)
func (h *UserHandler) ActivateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrBadRequest)
		return
	}

	err = h.userService.ActivateUser(c.Request.Context(), uint(id))
	if err != nil {
		switch err {
		case entities.ErrUserNotFound:
			c.JSON(http.StatusNotFound, dto.ErrUserNotFound)
		default:
			h.logger.Error("Activate user failed", zap.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, dto.ErrInternalServer)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User activated successfully"})
}

// DeactivateUser deactivates user account (admin only)
func (h *UserHandler) DeactivateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrBadRequest)
		return
	}

	err = h.userService.DeactivateUser(c.Request.Context(), uint(id))
	if err != nil {
		switch err {
		case entities.ErrUserNotFound:
			c.JSON(http.StatusNotFound, dto.ErrUserNotFound)
		default:
			h.logger.Error("Deactivate user failed", zap.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, dto.ErrInternalServer)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deactivated successfully"})
}
