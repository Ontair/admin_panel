package repository

import (
	"context"

	"github.com/ontair/admin-panel/internal/core/entities"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *entities.User) error
	// GetByID retrieves user by ID
	GetByID(ctx context.Context, id uint) (*entities.User, error)
	// GetByUsername retrieves user by username
	GetByUsername(ctx context.Context, username string) (*entities.User, error)
	// Update updates user data
	Update(ctx context.Context, user *entities.User) error
	// Delete deletes user by ID
	Delete(ctx context.Context, id uint) error
	// List retrieves list of users with pagination
	List(ctx context.Context, limit, offset int) ([]*entities.User, error)
	// Count returns total number of users
	Count(ctx context.Context) (int64, error)
	// GetByRole retrieves users by role
	GetByRole(ctx context.Context, role entities.Role) ([]*entities.User, error)
	// GetByRoles retrieves users by multiple roles
	GetByRoles(ctx context.Context, roles []entities.Role) ([]*entities.User, error)
	// UpdateLastLogin updates user's last login timestamp
	UpdateLastLogin(ctx context.Context, userID uint) error
}
