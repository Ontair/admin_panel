package services

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ontair/admin-panel/internal/core/entities"
	"github.com/ontair/admin-panel/internal/core/ports/repository"
	"github.com/ontair/admin-panel/internal/core/ports/service"
)

// AuthService implements AuthService interface
type AuthService struct {
	userRepo   repository.UserRepository
	jwtService service.JWTService
}

// NewAuthService creates new auth service
func NewAuthService(
	userRepo repository.UserRepository,
	jwtService service.JWTService,
) service.AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// Login authenticates user and returns tokens
func (s *AuthService) Login(ctx context.Context, req *service.LoginRequest) (*service.LoginResponse, error) {
	// Validate input
	if req.Username == "" || req.Password == "" {
		return nil, entities.ErrInvalidCredentials
	}

	// Get user by username
	user, err := s.getUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, entities.ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, entities.ErrUserDeactivated
	}

	// Verify password
	if !user.VerifyPassword(req.Password) {
		return nil, entities.ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	// Update last login
	user.UpdateLastLogin()
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Log error but don't fail login
		// TODO: Add proper logging
	}

	return &service.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		ExpiresIn:    15, // 15 minutes
	}, nil
}

// Register creates new user account
func (s *AuthService) Register(ctx context.Context, req *service.RegisterRequest) (*entities.User, error) {
	// Validate input
	if err := s.validateRegistrationRequest(req); err != nil {
		return nil, err
	}

	// Check if user already exists
	if _, err := s.userRepo.GetByUsername(ctx, req.Username); err == nil {
		return nil, entities.ErrUserAlreadyExists
	}

	// Create new user
	user := &entities.User{
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      getDefaultRole(req.Role),
		IsActive:  true,
	}

	// Set password
	if err := user.SetPassword(req.Password); err != nil {
		return nil, err
	}

	// Validate user entity
	if err := user.Validate(); err != nil {
		return nil, err
	}

	// Save user to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// RefreshToken generates new access token using refresh token
func (s *AuthService) RefreshToken(ctx context.Context, req *service.RefreshTokenRequest) (*service.LoginResponse, error) {
	// Validate refresh token
	token, err := s.jwtService.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, entities.ErrInvalidToken
	}

	// Extract user ID from token
	userID, ok := token.Claims.(jwt.MapClaims)["user_id"].(float64)
	if !ok {
		return nil, entities.ErrInvalidToken
	}

	// Get user from database
	user, err := s.userRepo.GetByID(ctx, uint(userID))
	if err != nil {
		return nil, entities.ErrUserNotFound
	}

	// Check if user is active
	if !user.IsActive {
		return nil, entities.ErrUserDeactivated
	}

	// Generate new tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.jwtService.GenerateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &service.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         user,
		ExpiresIn:    15, // 15 minutes
	}, nil
}

// Logout invalidates user session
func (s *AuthService) Logout(ctx context.Context, token string) error {
	// Parse token to get user ID
	parsedToken, err := s.jwtService.ParseAccessToken(token)
	if err != nil {
		return nil // Token already invalid
	}

	userID, ok := parsedToken.Claims.(jwt.MapClaims)["user_id"].(float64)
	if !ok {
		return nil
	}

	// For now, just return success - in production you might want to blacklist tokens
	_ = userID
	return nil
}

// ValidateToken validates JWT token
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*entities.User, error) {
	// Parse token
	parsedToken, err := s.jwtService.ParseAccessToken(token)
	if err != nil {
		return nil, entities.ErrInvalidToken
	}

	// Extract user ID
	userID, ok := parsedToken.Claims.(jwt.MapClaims)["user_id"].(float64)
	if !ok {
		return nil, entities.ErrInvalidToken
	}

	// Get user from database
	user, err := s.userRepo.GetByID(ctx, uint(userID))
	if err != nil {
		return nil, entities.ErrUserNotFound
	}

	// Check if user is active
	if !user.IsActive {
		return nil, entities.ErrUserDeactivated
	}

	return user, nil
}

// Helper methods

func (s *AuthService) getUserByUsername(ctx context.Context, username string) (*entities.User, error) {
	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, entities.ErrUserNotFound
	}
	return user, nil
}

func (s *AuthService) validateRegistrationRequest(req *service.RegisterRequest) error {
	if req.Username == "" || len(req.Username) < 3 {
		return entities.ErrInvalidUsername
	}

	if req.Password == "" || len(req.Password) < 8 {
		return entities.ErrPasswordTooShort
	}

	return nil
}

func getDefaultRole(role entities.Role) entities.Role {
	if role == "" {
		return entities.RoleUser
	}

	// Validate role
	switch role {
	case entities.RoleAdmin, entities.RoleManager, entities.RoleUser, entities.RoleGuest:
		return role
	default:
		return entities.RoleUser
	}
}
