package jwt

import (
	"fmt"
	"time"

	"github.com/ontair/admin-panel/internal/core/entities"
	"github.com/ontair/admin-panel/internal/core/ports/service"
	"github.com/ontair/admin-panel/internal/infra/config"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService handles JWT token operations
type JWTService struct {
	config        *config.Config
	accessClaims  map[string]interface{}
	refreshClaims map[string]interface{}
}

// Claims represent JWT claims
type Claims struct {
	jwt.RegisteredClaims
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Type     string `json:"type"` // "access" or "refresh"
}

// NewJWTService creates new JWT service
func NewJWTService(config *config.Config) *JWTService {
	return &JWTService{
		config: config,
		accessClaims: map[string]interface{}{
			"type": "access",
		},
		refreshClaims: map[string]interface{}{
			"type": "refresh",
		},
	}
}

// GenerateAccessToken generates access token for user
func (s *JWTService) GenerateAccessToken(user *entities.User) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "github.com/ontair/admin-panel",
			Subject:   fmt.Sprintf("%d", user.ID),
			Audience:  []string{"admin-panel-users"},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.config.JWT.AccessExpiry) * time.Minute)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:   user.ID,
		Username: user.Username,
		Role:     string(user.Role),
		Type:     "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.SecretKey))
}

// GenerateRefreshToken generates refresh token for user
func (s *JWTService) GenerateRefreshToken(user *entities.User) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "github.com/ontair/admin-panel",
			Subject:   fmt.Sprintf("%d", user.ID),
			Audience:  []string{"admin-panel-users"},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.config.JWT.RefreshExpiry) * time.Minute)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:   user.ID,
		Username: user.Username,
		Role:     string(user.Role),
		Type:     "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.RefreshSecret))
}

// ParseAccessToken parses and validates access token
func (s *JWTService) ParseAccessToken(tokenString string) (*jwt.Token, error) {
	return s.parseToken(tokenString, s.config.JWT.SecretKey, "access")
}

// ParseRefreshToken parses and validates refresh token
func (s *JWTService) ParseRefreshToken(tokenString string) (*jwt.Token, error) {
	return s.parseToken(tokenString, s.config.JWT.RefreshSecret, "refresh")
}

// parseToken parses token with specified secret and type
func (s *JWTService) parseToken(tokenString, secret, expectedType string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Validate token type
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if tokenType, ok := claims["type"].(string); !ok || tokenType != expectedType {
			return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, tokenType)
		}
	} else {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

// ExtractUserFromToken extracts user information from token
func (s *JWTService) ExtractUserFromToken(token *jwt.Token) (*service.UserInfo, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid user_id in token")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid username in token")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid role in token")
	}

	return &service.UserInfo{
		UserID:   uint(userIDFloat),
		Username: username,
		Role:     role,
	}, nil
}

// GetAccessTokenExpiry returns access token expiry in minutes
func (s *JWTService) GetAccessTokenExpiry() int {
	return s.config.JWT.AccessExpiry
}

// ValidateToken validates a token and returns claims
func (s *JWTService) ValidateToken(tokenString string) (*service.Claims, error) {
	token, err := s.ParseAccessToken(tokenString)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid user_id in token")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid username in token")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid role in token")
	}

	return &service.Claims{
		UserID:   uint(userIDFloat),
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", uint(userIDFloat)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.config.JWT.AccessExpiry) * time.Minute)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}, nil
}

// TokenUserInfo contains user information extracted from token
type TokenUserInfo struct {
	UserID   uint
	Username string
	Role     entities.Role
}

// GetUserID returns user ID from token
func (info *TokenUserInfo) GetUserID() uint {
	return info.UserID
}

// GetUsername returns username from token
func (info *TokenUserInfo) GetUsername() string {
	return info.Username
}

// GetRole returns user role from token
func (info *TokenUserInfo) GetRole() entities.Role {
	return info.Role
}

// IsAdmin checks if user is admin
func (info *TokenUserInfo) IsAdmin() bool {
	return info.Role == entities.RoleAdmin
}

// IsManagerOrHigher checks if user has manager privileges or higher
func (info *TokenUserInfo) IsManagerOrHigher() bool {
	return info.Role == entities.RoleAdmin || info.Role == entities.RoleManager
}

// HasRole checks if user has specific role
func (info *TokenUserInfo) HasRole(role entities.Role) bool {
	return info.Role == role
}
