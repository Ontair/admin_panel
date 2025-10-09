package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/ontair/admin-panel/internal/core/entities"
	"github.com/ontair/admin-panel/internal/core/ports/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository implements UserRepository interface using pgx
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates new user repository
func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	query := `
		INSERT INTO users (username, password, first_name, last_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query,
		user.Username,
		user.Password,
		user.FirstName,
		user.LastName,
		string(user.Role),
		user.IsActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID retrieves user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uint) (*entities.User, error) {
	query := `
		SELECT id, username, password, first_name, last_name, role, is_active, 
			   last_login, created_at, updated_at
		FROM users WHERE id = $1`

	var user entities.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.IsActive,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entities.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &user, nil
}

// GetByUsername retrieves user by username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	query := `
		SELECT id, username, password, first_name, last_name, role, is_active, 
			   last_login, created_at, updated_at
		FROM users WHERE username = $1`

	var user entities.User
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.IsActive,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entities.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &user, nil
}

// Update updates user data
func (r *UserRepository) Update(ctx context.Context, user *entities.User) error {
	query := `
		UPDATE users SET 
			username = $2, password = $3, first_name = $4, 
			last_name = $5, role = $6, is_active = $7, last_login = $8,
			updated_at = NOW()
		WHERE id = $1`

	cmdTag, err := r.db.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Password,
		user.FirstName,
		user.LastName,
		string(user.Role),
		user.IsActive,
		user.LastLogin,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return entities.ErrUserNotFound
	}

	return nil
}

// Delete deletes user by ID
func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	query := `DELETE FROM users WHERE id = $1`

	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return entities.ErrUserNotFound
	}

	return nil
}

// List retrieves list of users with pagination
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	query := `
		SELECT id, username, password, first_name, last_name, role, is_active, 
			   last_login, created_at, updated_at
		FROM users 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		var user entities.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Password,
			&user.FirstName,
			&user.LastName,
			&user.Role,
			&user.IsActive,
			&user.LastLogin,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return users, nil
}

// GetByRoles retrieves users by multiple roles
func (r *UserRepository) GetByRoles(ctx context.Context, roles []entities.Role) ([]*entities.User, error) {
	if len(roles) == 0 {
		return []*entities.User{}, nil
	}

	// Build query with IN clause
	placeholders := make([]string, len(roles))
	args := make([]interface{}, len(roles))
	for i, role := range roles {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = string(role)
	}

	query := fmt.Sprintf(`
		SELECT id, username, password, first_name, last_name, role, is_active, 
			   last_login, created_at, updated_at
		FROM users WHERE role IN (%s)
		ORDER BY created_at DESC`, strings.Join(placeholders, ","))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by roles: %w", err)
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		var user entities.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Password,
			&user.FirstName,
			&user.LastName,
			&user.Role,
			&user.IsActive,
			&user.LastLogin,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return users, nil
}

// Count returns total number of users
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM users`

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}

// GetByRole retrieves users by role
func (r *UserRepository) GetByRole(ctx context.Context, role entities.Role) ([]*entities.User, error) {
	query := `
		SELECT id, username, password, first_name, last_name, role, is_active, 
			   last_login, created_at, updated_at
		FROM users WHERE role = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, string(role))
	if err != nil {
		return nil, fmt.Errorf("failed to get users by role: %w", err)
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		var user entities.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Password,
			&user.FirstName,
			&user.LastName,
			&user.Role,
			&user.IsActive,
			&user.LastLogin,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return users, nil
}

// UpdateLastLogin updates user's last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uint) error {
	query := `UPDATE users SET last_login = NOW() WHERE id = $1`

	cmdTag, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return entities.ErrUserNotFound
	}

	return nil
}
