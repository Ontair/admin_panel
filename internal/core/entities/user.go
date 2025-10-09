package entities

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user entity in the domain
type User struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	Username  string     `json:"username" gorm:"unique;not null"`
	Password  string     `json:"-" gorm:"not null"` // Hidden in JSON
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Role      Role       `json:"role" gorm:"type:varchar(20);default:'user'"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	LastLogin *time.Time `json:"last_login"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Role represents user roles
type Role string

const (
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleUser    Role = "user"
	RoleGuest   Role = "guest"
)

// HasRole checks if user has specific role
func (u *User) HasRole(role Role) bool {
	return u.Role == role
}

// IsAdmin checks if user is admin
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsManagerOrHigher checks if user has manager or admin role
func (u *User) IsManagerOrHigher() bool {
	return u.Role == RoleAdmin || u.Role == RoleManager
}

// SetPassword hashes the password
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// VerifyPassword verifies the password
func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// UpdateLastLogin updates the last login time
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLogin = &now
}

// Validate validates user data
func (u *User) Validate() error {
	if u.Username == "" {
		return ErrInvalidUsername
	}
	if len(u.Password) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}
