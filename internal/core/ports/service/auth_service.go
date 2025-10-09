package service

import (
	"context"

	"github.com/ontair/admin-panel/internal/core/entities"
)

// LoginRequest represents login request data
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents login response data
type LoginResponse struct {
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	User         *entities.User `json:"user"`
	ExpiresIn    int            `json:"expires_in"`
}

// RegisterRequest represents registration request data
type RegisterRequest struct {
	Username  string        `json:"username" validate:"required,min=3"`
	Password  string        `json:"password" validate:"required,min=8"`
	FirstName string        `json:"first_name"`
	LastName  string        `json:"last_name"`
	Role      entities.Role `json:"role"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AuthService defines authentication service interface
type AuthService interface {
	// Login authenticates user and returns tokens
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	// Register creates new user account
	Register(ctx context.Context, req *RegisterRequest) (*entities.User, error)
	// RefreshToken generates new access token using refresh token
	RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*LoginResponse, error)
	// Logout invalidates user session
	Logout(ctx context.Context, token string) error
	// ValidateToken validates JWT token
	ValidateToken(ctx context.Context, token string) (*entities.User, error)
}
