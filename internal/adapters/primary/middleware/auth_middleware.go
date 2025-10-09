package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ontair/admin-panel/internal/core/entities"
	"github.com/ontair/admin-panel/internal/core/ports/service"
	"go.uber.org/zap"
)

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	jwtService    service.JWTService
	logger        service.Logger
	cookieService service.CookieService
	authService   service.AuthService
}

// NewAuthMiddleware creates new auth middleware
func NewAuthMiddleware(jwtService service.JWTService, logger service.Logger, cookieService service.CookieService, authService service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService:    jwtService,
		logger:        logger,
		cookieService: cookieService,
		authService:   authService,
	}
}

// RequireAuth middleware that requires authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.logger.Info("RequireAuth middleware called")
		token, err := m.extractToken(c)
		if err != nil {
			m.logger.Info("Failed to extract token", zap.String("error", err.Error()))
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized",
				"message": "Invalid or missing token",
				"details": "Please provide a valid authentication token",
			})
			c.Abort()
			return
		}

		// Validate token
		parsedToken, err := m.jwtService.ParseAccessToken(token)
		if err != nil {
			// Check if token is expired and try to refresh
			if m.isTokenExpiredError(err) {
				m.logger.Info("Access token expired, attempting refresh")
				if m.attemptTokenRefresh(c) {
					// Token refresh successful, continue with the request
					c.Next()
					return
				}
			}

			m.logger.Info("Invalid token", zap.String("error", err.Error()))
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized",
				"message": "Invalid token",
				"details": "Token validation failed",
			})
			c.Abort()
			return
		}

		// Extract user info from token
		userInfo, err := m.jwtService.ExtractUserFromToken(parsedToken)
		if err != nil {
			m.logger.Error("Failed to extract user from token", zap.String("error", err.Error()))
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized",
				"message": "Invalid token data",
				"details": "Token contains invalid user information",
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", userInfo.UserID)
		c.Set("username", userInfo.Username)
		c.Set("role", userInfo.Role) // Keep as string for consistency
		c.Set("user_info", userInfo)

		m.logger.Info("User authenticated successfully", zap.String("username", userInfo.Username), zap.String("role", userInfo.Role))
		c.Next()
	}
}

// RequireRole middleware that requires specific role
func (m *AuthMiddleware) RequireRole(role entities.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized",
				"message": "User role not found",
			})
			c.Abort()
			return
		}

		userRoleStr, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Internal Server Error",
				"message": "Invalid user role format",
			})
			c.Abort()
			return
		}

		if entities.Role(userRoleStr) != role {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Forbidden",
				"message": fmt.Sprintf("Required role: %s", role),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin middleware that requires admin role
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole(entities.RoleAdmin)
}

// RequireManagerOrHigher middleware that requires manager or admin role
func (m *AuthMiddleware) RequireManagerOrHigher() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized",
				"message": "User role not found",
			})
			c.Abort()
			return
		}

		userRoleStr, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Internal Server Error",
				"message": "Invalid user role format",
			})
			c.Abort()
			return
		}

		role := entities.Role(userRoleStr)
		if role != entities.RoleAdmin && role != entities.RoleManager {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Forbidden",
				"message": "Required role: manager or admin",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Helper methods

func (m *AuthMiddleware) extractToken(c *gin.Context) (string, error) {
	// Try to get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1], nil
		}
	}

	// Try to get token from cookie
	token, err := m.cookieService.GetTokenFromRequest(c)
	if err == nil && token != "" {
		return token, nil
	}

	return "", fmt.Errorf("no token found")
}

func (m *AuthMiddleware) isTokenExpiredError(err error) bool {
	// Check if error indicates token expiration
	return err != nil && (err.Error() == "token is expired" || err.Error() == "Token is expired")
}

func (m *AuthMiddleware) attemptTokenRefresh(c *gin.Context) bool {
	// Try to get refresh token from cookie
	refreshToken, err := m.cookieService.GetRefreshToken(c)
	if err != nil {
		m.logger.Info("Failed to get refresh token", zap.String("error", err.Error()))
		return false
	}

	// Attempt to refresh token
	refreshReq := &service.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}

	response, err := m.authService.RefreshToken(c.Request.Context(), refreshReq)
	if err != nil {
		m.logger.Info("Token refresh failed", zap.String("error", err.Error()))
		return false
	}

	// Set new cookies
	m.cookieService.SetAuthCookies(c, response.AccessToken, response.RefreshToken)

	// Parse new access token to get user info
	parsedToken, err := m.jwtService.ParseAccessToken(response.AccessToken)
	if err != nil {
		m.logger.Info("New token parsing failed", zap.String("error", err.Error()))
		return false
	}

	// Extract user info from new token
	userInfo, err := m.jwtService.ExtractUserFromToken(parsedToken)
	if err != nil {
		m.logger.Info("Failed to extract user from new token", zap.String("error", err.Error()))
		return false
	}

	// Set user information in context
	c.Set("user_id", userInfo.UserID)
	c.Set("username", userInfo.Username)
	c.Set("role", userInfo.Role) // Keep as string for consistency
	c.Set("user_info", userInfo)

	m.logger.Info("Token refreshed successfully", zap.String("username", userInfo.Username))
	return true
}
