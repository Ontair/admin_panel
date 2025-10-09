package service

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/ontair/admin-panel/internal/core/entities"
)

// JWTService defines the interface for JWT operations
type JWTService interface {
	GenerateAccessToken(user *entities.User) (string, error)
	GenerateRefreshToken(user *entities.User) (string, error)
	ParseAccessToken(tokenString string) (*jwt.Token, error)
	ParseRefreshToken(tokenString string) (*jwt.Token, error)
	ExtractUserFromToken(token *jwt.Token) (*UserInfo, error)
	ValidateToken(tokenString string) (*Claims, error)
}

// UserInfo contains user information extracted from JWT
type UserInfo struct {
	UserID   uint
	Username string
	Role     string
}

// Claims represents JWT claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}
