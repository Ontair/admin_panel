package cookie

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ontair/admin-panel/internal/core/ports/service"
)

// CookieService implements port.CookieService interface
type CookieService struct {
	accessTokenName  string
	refreshTokenName string
	domain           string
	secure           bool
	httpOnly         bool
	sameSite         http.SameSite
}

// NewCookieService creates new cookie service
func NewCookieService() service.CookieService {
	return &CookieService{
		accessTokenName:  "access_token",
		refreshTokenName: "refresh_token",
		domain:           "",    // Use default domain
		secure:           false, // Set to true in production with HTTPS
		httpOnly:         true,
		sameSite:         http.SameSiteStrictMode,
	}
}

// SetAuthCookies sets access and refresh token cookies
func (s *CookieService) SetAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	// Set access token cookie (short-lived)
	c.SetCookie(
		s.accessTokenName,
		accessToken,
		15*60, // 15 minutes
		"/",
		s.domain,
		s.secure,
		s.httpOnly,
	)

	// Set refresh token cookie (long-lived)
	c.SetCookie(
		s.refreshTokenName,
		refreshToken,
		7*24*60*60, // 7 days
		"/",
		s.domain,
		s.secure,
		s.httpOnly,
	)
}

// GetAccessToken retrieves access token from cookie
func (s *CookieService) GetAccessToken(c *gin.Context) (string, error) {
	token, err := c.Cookie(s.accessTokenName)
	if err != nil {
		return "", errors.New("access token not found in cookie")
	}
	return token, nil
}

// GetRefreshToken retrieves refresh token from cookie
func (s *CookieService) GetRefreshToken(c *gin.Context) (string, error) {
	token, err := c.Cookie(s.refreshTokenName)
	if err != nil {
		return "", errors.New("refresh token not found in cookie")
	}
	return token, nil
}

// GetTokenFromRequest retrieves token from request (cookie or header)
func (s *CookieService) GetTokenFromRequest(c *gin.Context) (string, error) {
	// Try to get token from cookie first
	if token, err := s.GetAccessToken(c); err == nil {
		return token, nil
	}

	// Fallback to Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("no token found in request")
	}

	// Check if it starts with "Bearer "
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:], nil
	}

	return "", errors.New("invalid authorization header format")
}

// ClearAuthCookies clears authentication cookies
func (s *CookieService) ClearAuthCookies(c *gin.Context) {
	// Clear access token cookie
	c.SetCookie(
		s.accessTokenName,
		"",
		-1, // Expire immediately
		"/",
		s.domain,
		s.secure,
		s.httpOnly,
	)

	// Clear refresh token cookie
	c.SetCookie(
		s.refreshTokenName,
		"",
		-1, // Expire immediately
		"/",
		s.domain,
		s.secure,
		s.httpOnly,
	)
}
