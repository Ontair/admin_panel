package service

import (
	"context"

	"github.com/ontair/admin-panel/internal/core/entities"
)

// CreateUserRequest represents user creation request
type CreateUserRequest struct {
	Username  string        `json:"username" validate:"required,min=3"`
	Password  string        `json:"password" validate:"required,min=8"`
	FirstName string        `json:"first_name"`
	LastName  string        `json:"last_name"`
	Role      entities.Role `json:"role"`
	IsActive  bool          `json:"is_active"`
}

// UpdateUserRequest represents user update request
type UpdateUserRequest struct {
	Username  *string        `json:"username"`
	FirstName *string        `json:"first_name"`
	LastName  *string        `json:"last_name"`
	Role      *entities.Role `json:"role"`
	IsActive  *bool          `json:"is_active"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// ResetPasswordRequest represents password reset request
type ResetPasswordRequest struct {
	Username string `json:"username" validate:"required"`
}

// ConfirmPasswordResetRequest represents password reset confirmation
type ConfirmPasswordResetRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ListUsersRequest represents user listing request with filters and pagination
type ListUsersRequest struct {
	Limit    int           `query:"limit"`
	Offset   int           `query:"offset"`
	Role     entities.Role `query:"role"`
	IsActive *bool         `query:"is_active"`
	Search   string        `query:"search"`
}

// ListUsersResponse represents paginated users response
type ListUsersResponse struct {
	Users  []*entities.User `json:"users"`
	Total  int64            `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

// UserService defines user management service interface
type UserService interface {
	// CreateUser creates a new user (admin only)
	CreateUser(ctx context.Context, req *CreateUserRequest) (*entities.User, error)
	// GetUser retrieves user by ID
	GetUser(ctx context.Context, id uint) (*entities.User, error)
	// GetCurrentUser retrieves current authenticated user
	GetCurrentUser(ctx context.Context, userID uint) (*entities.User, error)
	// UpdateUser updates user data
	UpdateUser(ctx context.Context, id uint, req *UpdateUserRequest) (*entities.User, error)
	// DeleteUser deletes user by ID (admin only)
	DeleteUser(ctx context.Context, id uint) error
	// ListUsers retrieves paginated list of users
	ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error)
	// ListUsersForManager retrieves paginated list of users for manager (only user and guest roles)
	ListUsersForManager(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error)
	// ChangePassword allows user to change their password
	ChangePassword(ctx context.Context, userID uint, req *ChangePasswordRequest) error
	// ResetPassword initiates password reset process
	ResetPassword(ctx context.Context, req *ResetPasswordRequest) error
	// ConfirmPasswordReset confirms password reset with token
	ConfirmPasswordReset(ctx context.Context, req *ConfirmPasswordResetRequest) error
	// ActivateUser activates user account (admin only)
	ActivateUser(ctx context.Context, id uint) error
	// DeactivateUser deactivates user account (admin only)
	DeactivateUser(ctx context.Context, id uint) error
}
