package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ontair/admin-panel/internal/core/entities"
	"github.com/ontair/admin-panel/internal/infra/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DatabaseService handles database operations with pgx
type DatabaseService struct {
	db     *pgxpool.Pool
	config *config.Config
}

// NewDatabaseService creates new database service
func NewDatabaseService(cfg *config.Config) (*DatabaseService, error) {
	// Parse configuration
	dbURL := cfg.GetPostgresURL()

	// Configure connection pool
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure pool settings
	config.MaxConns = 100
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30

	// Connect to database
	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	service := &DatabaseService{
		db:     db,
		config: cfg,
	}

	// Test connection
	if err := service.Health(); err != nil {
		db.Close()
		return nil, fmt.Errorf("database health check failed: %w", err)
	}

	// Initialize database migrations
	if err := service.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return service, nil
}

// GetPool returns database connection pool
func (s *DatabaseService) GetPool() *pgxpool.Pool {
	return s.db
}

// Close closes database connection
func (s *DatabaseService) Close() error {
	s.db.Close()
	return nil
}

// Health checks database health
func (s *DatabaseService) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result int
	err := s.db.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}

// migrate runs database migrations
func (s *DatabaseService) migrate() error {
	log.Println("Running database migrations...")

	ctx := context.Background()

	// Create users table
	if err := s.createUsersTable(ctx); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}
	log.Println("Users table created successfully")

	// Create indexes
	if err := s.createIndexes(ctx); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}
	log.Println("Database indexes created successfully")

	// Seed data if needed
	if err := s.seedData(ctx); err != nil {
		return fmt.Errorf("failed to seed data: %w", err)
	}
	log.Println("Initial data seeded successfully")

	log.Println("Database migrations completed successfully")
	return nil
}

// createUsersTable creates users table
func (s *DatabaseService) createUsersTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			first_name VARCHAR(50),
			last_name VARCHAR(50),
			role VARCHAR(20) DEFAULT 'user' NOT NULL CHECK (role IN ('admin', 'manager', 'user', 'guest')),
			is_active BOOLEAN DEFAULT true NOT NULL,
			last_login TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`

	_, err := s.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Verify table was created
	var tableExists bool
	err = s.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'users'
		)`).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to verify users table: %w", err)
	}

	if !tableExists {
		return fmt.Errorf("users table was not created")
	}

	log.Println("Users table verified successfully")
	return nil
}

// createIndexes creates database indexes
func (s *DatabaseService) createIndexes(ctx context.Context) error {
	indexes := []string{
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username)",
		"CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)",
		"CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active)",
	}

	for _, idx := range indexes {
		if _, err := s.db.Exec(ctx, idx); err != nil {
			log.Printf("Warning: Index creation failed: %v", err)
		}
	}

	return nil
}

// seedData seeds initial data if needed
func (s *DatabaseService) seedData(ctx context.Context) error {
	// Check if admin user exists
	var adminCount int
	err := s.db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&adminCount)
	if err != nil {
		return fmt.Errorf("failed to count admin users: %w", err)
	}

	// Create default admin if no admin exists
	if adminCount == 0 {
		log.Println("Creating default admin user...")

		// Create admin user with hashed password
		admin := &entities.User{
			Username:  "admin",
			FirstName: "Admin",
			LastName:  "User",
			Role:      entities.RoleAdmin,
			IsActive:  true,
		}

		// Set default password
		if err := admin.SetPassword("admin123"); err != nil {
			return fmt.Errorf("failed to set admin password: %w", err)
		}

		// Insert admin user
		query := `
			INSERT INTO users (username, password, first_name, last_name, role, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`

		_, err := s.db.Exec(ctx, query,
			admin.Username,
			admin.Password,
			admin.FirstName,
			admin.LastName,
			admin.Role,
			admin.IsActive,
		)

		if err != nil {
			log.Printf("Warning: Failed to create default admin user: %v", err)
		} else {
			log.Println("Default admin user created successfully")
			log.Println("Login credentials:")
			log.Println("  Username: admin")
			log.Println("  Password: admin123")
		}
	}

	return nil
}
