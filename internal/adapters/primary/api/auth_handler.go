package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ontair/admin-panel/internal/core/dto"
	"github.com/ontair/admin-panel/internal/core/entities"
	"github.com/ontair/admin-panel/internal/core/ports/service"
	"go.uber.org/zap"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService   service.AuthService
	logger        service.Logger
	cookieService service.CookieService
	jwtService    service.JWTService
}

// NewAuthHandler creates new auth handler
func NewAuthHandler(authService service.AuthService, logger service.Logger, cookieService service.CookieService, jwtService service.JWTService) *AuthHandler {
	return &AuthHandler{
		authService:   authService,
		logger:        logger,
		cookieService: cookieService,
		jwtService:    jwtService,
	}
}

// RegisterPublicRoutes registers public auth routes (no authentication required)
func (h *AuthHandler) RegisterPublicRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshToken)
		auth.POST("/logout", h.Logout)
	}
}

// RegisterProtectedRoutes registers protected auth routes (authentication required)
func (h *AuthHandler) RegisterProtectedRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.GET("/profile", h.GetProfile)
	}
}

// RegisterManagerRoutes registers manager+ auth routes (manager and admin only)
func (h *AuthHandler) RegisterManagerRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", h.Register) // Only manager+ can register users
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var loginDTO dto.LoginDTO
	if err := c.ShouldBindJSON(&loginDTO); err != nil {
		h.logger.Error("Invalid login request", zap.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Convert to service request
	loginReq := &service.LoginRequest{
		Username: loginDTO.Username,
		Password: loginDTO.Password,
	}

	// Authenticate user
	response, err := h.authService.Login(c.Request.Context(), loginReq)
	if err != nil {
		switch err {
		case entities.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized",
				"message": "Invalid credentials",
				"details": "Username or password is incorrect",
			})
		case entities.ErrUserNotFound:
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized",
				"message": "Invalid credentials",
				"details": "Username or password is incorrect",
			})
		case entities.ErrUserDeactivated:
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Forbidden",
				"message": "Account is deactivated",
				"details": "Your account has been deactivated. Please contact an administrator.",
			})
		default:
			// Log only unexpected errors
			h.logger.Error("Login failed", zap.String("username", loginDTO.Username), zap.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Internal Server Error",
				"message": "Login failed",
			})
		}
		return
	}

	// Set authentication cookies
	h.cookieService.SetAuthCookies(c, response.AccessToken, response.RefreshToken)

	// Convert to DTO (without tokens for security)
	authResponse := dto.AuthResponseDTO{
		User:      dto.ToUserDTO(response.User),
		ExpiresIn: response.ExpiresIn,
	}

	h.logger.Info("User logged in successfully", zap.String("username", loginDTO.Username))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    authResponse,
		"message": "Login successful",
	})
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var registerDTO dto.RegisterDTO
	if err := c.ShouldBindJSON(&registerDTO); err != nil {
		h.logger.Error("Invalid registration request", zap.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Convert to service request
	registerReq := &service.RegisterRequest{
		Username:  registerDTO.Username,
		Password:  registerDTO.Password,
		FirstName: registerDTO.FirstName,
		LastName:  registerDTO.LastName,
		Role:      entities.RoleUser, // Default role for registration
	}

	// Register user
	user, err := h.authService.Register(c.Request.Context(), registerReq)
	if err != nil {
		switch err {
		case entities.ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Conflict",
				"message": "User already exists",
			})
		case entities.ErrInvalidUsername, entities.ErrPasswordTooShort:
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": err.Error(),
			})
		default:
			// Log only unexpected errors
			h.logger.Error("Registration failed", zap.String("username", registerDTO.Username), zap.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Registration failed",
			})
		}
		return
	}

	h.logger.Info("User registered successfully", zap.String("username", user.Username))

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User registered successfully",
		"data":    dto.ToUserDTO(user),
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Get refresh token from cookie or request body
	refreshToken, err := h.cookieService.GetRefreshToken(c)
	if err != nil {
		// Fallback to request body
		var refreshReq dto.JWTResponseDTO
		if err := c.ShouldBindJSON(&refreshReq); err != nil {
			h.logger.Error("Invalid refresh token request", zap.String("error", err.Error()))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": "No refresh token found",
			})
			return
		}
		refreshToken = refreshReq.RefreshToken
	}

	// Convert DTO to service request
	serviceReq := &service.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}

	// Call service
	response, err := h.authService.RefreshToken(c.Request.Context(), serviceReq)
	if err != nil {
		switch err {
		case entities.ErrInvalidToken, entities.ErrTokenExpired:
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized",
				"message": "Invalid or expired refresh token",
				"details": "Please login again",
			})
		case entities.ErrUserNotFound:
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized",
				"message": "User not found",
				"details": "User account no longer exists",
			})
		case entities.ErrUserDeactivated:
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Forbidden",
				"message": "Account is deactivated",
				"details": "Your account has been deactivated. Please contact an administrator.",
			})
		default:
			h.logger.Error("Token refresh failed", zap.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Internal Server Error",
				"message": "Token refresh failed",
			})
		}
		return
	}

	// Set new cookies
	h.cookieService.SetAuthCookies(c, response.AccessToken, response.RefreshToken)

	// Convert to DTO (without tokens for security)
	authResponse := dto.AuthResponseDTO{
		User:      dto.ToUserDTO(response.User),
		ExpiresIn: response.ExpiresIn,
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Token refreshed successfully",
		"data":    authResponse,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get token from cookie or header (for backward compatibility)
	token, err := h.cookieService.GetTokenFromRequest(c)
	if err != nil {
		// No token means user is already logged out
		h.logger.Info("Logout attempted but no token found - user already logged out")
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Already logged out",
		})
		return
	}

	// Extract user info from token for logging BEFORE logout
	var userID uint
	var username string
	parsedToken, err := h.jwtService.ParseAccessToken(token)
	if err == nil {
		userInfo, err := h.jwtService.ExtractUserFromToken(parsedToken)
		if err == nil {
			userID = userInfo.UserID
			username = userInfo.Username
		}
	}

	// Logout user
	err = h.authService.Logout(c.Request.Context(), token)
	if err != nil {
		// Log only unexpected errors
		h.logger.Error("Logout failed", zap.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Logout failed",
		})
		return
	}

	// Clear authentication cookies
	h.cookieService.ClearAuthCookies(c)

	// Log successful logout with user info if available
	if userID != 0 && username != "" {
		h.logger.Info("User logged out successfully", zap.Uint("userID", userID), zap.String("username", username))
	} else {
		h.logger.Info("User logged out successfully")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}

// GetProfile returns current user profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	id, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Invalid user ID",
		})
		return
	}

	// For this endpoint, we'll just return the token data
	// In a real implementation, you might want to fetch fresh user data from database
	userRole, _ := c.Get("role")
	username, _ := c.Get("username")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":       id,
			"username": username,
			"role":     userRole,
		},
	})
}
