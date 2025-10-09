package service

import "github.com/gin-gonic/gin"

// CookieService defines the interface for cookie operations
type CookieService interface {
	SetAuthCookies(c *gin.Context, accessToken, refreshToken string)
	ClearAuthCookies(c *gin.Context)
	GetAccessToken(c *gin.Context) (string, error)
	GetRefreshToken(c *gin.Context) (string, error)
	GetTokenFromRequest(c *gin.Context) (string, error)
}
