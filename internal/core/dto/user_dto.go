package dto

import (
	"time"

	"github.com/ontair/admin-panel/internal/core/entities"
)

// UserDTO represents user data transfer object
type UserDTO struct {
	ID        uint       `json:"id"`
	Username  string     `json:"username"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Role      string     `json:"role"`
	IsActive  bool       `json:"is_active"`
	LastLogin *time.Time `json:"last_login"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// UserCreateDTO represents user creation DTO
type UserCreateDTO struct {
	Username  string `json:"username" validate:"required,min=3"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
}

// UserUpdateDTO represents user update DTO
type UserUpdateDTO struct {
	Username  *string `json:"username"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Role      *string `json:"role"`
	IsActive  *bool   `json:"is_active"`
}

// LoginDTO represents login DTO
type LoginDTO struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RegisterDTO represents registration DTO
type RegisterDTO struct {
	Username  string `json:"username" validate:"required,min=3"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// ChangePasswordDTO represents password change DTO
type ChangePasswordDTO struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// ResetPasswordDTO represents password reset DTO
type ResetPasswordDTO struct {
	Username string `json:"username" validate:"required"`
}

// JWTResponseDTO represents JWT response DTO
type JWTResponseDTO struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	User         UserDTO `json:"user"`
	ExpiresIn    int     `json:"expires_in"`
}

// AuthResponseDTO represents authentication response DTO (for cookies)
type AuthResponseDTO struct {
	User      UserDTO `json:"user"`
	ExpiresIn int     `json:"expires_in"`
}

// ToUserDTO converts domain user entity to DTO
func ToUserDTO(user *entities.User) UserDTO {
	return UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      string(user.Role),
		IsActive:  user.IsActive,
		LastLogin: user.LastLogin,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
