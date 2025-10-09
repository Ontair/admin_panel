package services

import (
	"context"
	"strings"

	"github.com/ontair/admin-panel/internal/core/entities"
	"github.com/ontair/admin-panel/internal/core/ports/repository"
	"github.com/ontair/admin-panel/internal/core/ports/service"
)

// UserService implements UserService interface
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService creates new user service
func NewUserService(userRepo repository.UserRepository) service.UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user (admin only)
func (s *UserService) CreateUser(ctx context.Context, req *service.CreateUserRequest) (*entities.User, error) {
	// Validate input
	if err := s.validateCreateUserRequest(req); err != nil {
		return nil, err
	}

	// Check if user already exists
	if _, err := s.userRepo.GetByUsername(ctx, req.Username); err == nil {
		return nil, entities.ErrUserAlreadyExists
	}

	// Create new user
	user := &entities.User{
		Username:  req.Username,
		Password:  "", // Will be set below
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      s.getValidRole(req.Role),
		IsActive:  req.IsActive,
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

// GetUser retrieves user by ID
func (s *UserService) GetUser(ctx context.Context, id uint) (*entities.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, entities.ErrUserNotFound
	}

	return user, nil
}

// GetCurrentUser retrieves current authenticated user
func (s *UserService) GetCurrentUser(ctx context.Context, userID uint) (*entities.User, error) {
	return s.GetUser(ctx, userID)
}

// UpdateUser updates user data
func (s *UserService) UpdateUser(ctx context.Context, id uint, req *service.UpdateUserRequest) (*entities.User, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, entities.ErrUserNotFound
	}

	// Validate update request
	if err := s.validateUpdateUserRequest(req); err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Username != nil {
		// Check if new username is available
		if existingUser, err := s.userRepo.GetByUsername(ctx, *req.Username); err == nil && existingUser.ID != user.ID {
			return nil, entities.ErrUserAlreadyExists
		}
		user.Username = *req.Username
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}

	if req.LastName != nil {
		user.LastName = *req.LastName
	}

	if req.Role != nil {
		user.Role = s.getValidRole(*req.Role)
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	// Validate updated user
	if err := user.Validate(); err != nil {
		return nil, err
	}

	// Save updates
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes user by ID (admin only)
func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	// Check if user exists
	if _, err := s.userRepo.GetByID(ctx, id); err != nil {
		return entities.ErrUserNotFound
	}

	// Delete user
	return s.userRepo.Delete(ctx, id)
}

// ListUsers retrieves paginated list of users
func (s *UserService) ListUsers(ctx context.Context, req *service.ListUsersRequest) (*service.ListUsersResponse, error) {
	// Set default pagination values
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20 // Default limit
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	var users []*entities.User
	var total int64
	var err error

	// Apply filters
	if req.Role != "" && req.Search != "" {
		// Search by role and text
		users, total, err = s.searchUsersByRoleAndText(ctx, req.Role, req.Search, limit, offset)
	} else if req.Role != "" {
		// Filter by role only
		users, err = s.userRepo.GetByRole(ctx, req.Role)
		if err == nil {
			total = int64(len(users))
			users = s.paginateUsers(users, limit, offset)
		}
	} else if req.Search != "" {
		// Search by text only
		users, total, err = s.searchUsersByText(ctx, req.Search, limit, offset)
	} else {
		// Get all users
		users, err = s.userRepo.List(ctx, limit, offset)
		if err == nil {
			total, err = s.userRepo.Count(ctx)
		}
	}

	if err != nil {
		return nil, err
	}

	// Filter by IsActive if specified
	if req.IsActive != nil {
		users = s.filterUsersByActiveStatus(users, *req.IsActive)
	}

	return &service.ListUsersResponse{
		Users:  users,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// ListUsersForManager retrieves paginated list of users for manager (only user and guest roles)
func (s *UserService) ListUsersForManager(ctx context.Context, req *service.ListUsersRequest) (*service.ListUsersResponse, error) {
	// Set default pagination values
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20 // Default limit
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	var users []*entities.User
	var total int64
	var err error

	// Manager can only see user and guest roles
	requestedRole := req.Role
	if requestedRole != "" && requestedRole != entities.RoleUser && requestedRole != entities.RoleGuest {
		// Return empty result for invalid roles
		return &service.ListUsersResponse{
			Users:  []*entities.User{},
			Total:  0,
			Limit:  limit,
			Offset: offset,
		}, nil
	}

	// If no specific role requested, get both user and guest roles
	if requestedRole == "" {
		// Get users with user and guest roles
		users, err = s.userRepo.GetByRoles(ctx, []entities.Role{entities.RoleUser, entities.RoleGuest})
		if err != nil {
			return nil, err
		}

		total = int64(len(users))

		// Apply search filter if specified
		if req.Search != "" {
			users = s.filterUsersByText(users, req.Search)
			total = int64(len(users))
		}

		// Apply pagination
		users = s.paginateUsers(users, limit, offset)
	} else {
		// Get specific role (only user or guest allowed)
		users, err = s.userRepo.GetByRole(ctx, requestedRole)
		if err != nil {
			return nil, err
		}

		total = int64(len(users))

		// Apply search filter if specified
		if req.Search != "" {
			users = s.filterUsersByText(users, req.Search)
			total = int64(len(users))
		}

		// Apply pagination
		users = s.paginateUsers(users, limit, offset)
	}

	if err != nil {
		return nil, err
	}

	// Filter by IsActive if specified
	if req.IsActive != nil {
		users = s.filterUsersByActiveStatus(users, *req.IsActive)
	}

	return &service.ListUsersResponse{
		Users:  users,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// ChangePassword allows user to change their password
func (s *UserService) ChangePassword(ctx context.Context, userID uint, req *service.ChangePasswordRequest) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return entities.ErrUserNotFound
	}

	// Verify current password
	if !user.VerifyPassword(req.CurrentPassword) {
		return entities.ErrInvalidCredentials
	}

	// Validate new password
	if len(req.NewPassword) < 8 {
		return entities.ErrPasswordTooShort
	}

	// Set new password
	if err := user.SetPassword(req.NewPassword); err != nil {
		return err
	}

	// Save updated user
	return s.userRepo.Update(ctx, user)
}

// ResetPassword initiates password reset process
func (s *UserService) ResetPassword(ctx context.Context, req *service.ResetPasswordRequest) error {
	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		// Don't reveal if user exists or not for security
		return nil
	}

	// TODO: Implement password reset token generation and email sending
	// For now, just validate that user exists
	_ = user
	// This would typically involve:
	// 1. Generate reset token
	// 2. Store token with expiration
	// 3. Send email with reset link

	return nil
}

// ConfirmPasswordReset confirms password reset with token
func (s *UserService) ConfirmPasswordReset(ctx context.Context, req *service.ConfirmPasswordResetRequest) error {
	// TODO: Implement password reset confirmation
	// For now, just validate input
	if req.Token == "" || len(req.NewPassword) < 8 {
		return entities.ErrPasswordTooShort
	}

	// This would typically involve:
	// 1. Validate reset token and get user ID
	// 2. Set new password
	// 3. Invalidate token

	return nil
}

// ActivateUser activates user account (admin only)
func (s *UserService) ActivateUser(ctx context.Context, id uint) error {
	return s.toggleUserActiveStatus(ctx, id, true)
}

// DeactivateUser deactivates user account (admin only)
func (s *UserService) DeactivateUser(ctx context.Context, id uint) error {
	return s.toggleUserActiveStatus(ctx, id, false)
}

// Private helper methods

func (s *UserService) validateCreateUserRequest(req *service.CreateUserRequest) error {
	if req.Username == "" || len(req.Username) < 3 {
		return entities.ErrInvalidUsername
	}

	if req.Password == "" || len(req.Password) < 8 {
		return entities.ErrPasswordTooShort
	}

	return nil
}

func (s *UserService) validateUpdateUserRequest(req *service.UpdateUserRequest) error {
	if req.Username != nil && (*req.Username == "" || len(*req.Username) < 3) {
		return entities.ErrInvalidUsername
	}

	return nil
}

func (s *UserService) getValidRole(role entities.Role) entities.Role {
	switch role {
	case entities.RoleAdmin, entities.RoleManager, entities.RoleUser, entities.RoleGuest:
		return role
	default:
		return entities.RoleUser
	}
}

func (s *UserService) searchUsersByRoleAndText(ctx context.Context, role entities.Role, search string, limit, offset int) ([]*entities.User, int64, error) {
	// TODO: Implement database-specific search
	// For now, get by role and filter in memory
	users, err := s.userRepo.GetByRole(ctx, role)
	if err != nil {
		return nil, 0, err
	}

	// Filter by search term
	filteredUsers := s.filterUsersByText(users, search)
	total := int64(len(filteredUsers))

	return s.paginateUsers(filteredUsers, limit, offset), total, nil
}

func (s *UserService) searchUsersByText(ctx context.Context, search string, limit, offset int) ([]*entities.User, int64, error) {
	// TODO: Implement database-specific search
	// For now, get all users and filter in memory
	allUsers, err := s.userRepo.List(ctx, 10000, 0) // Get large batch for search
	if err != nil {
		return nil, 0, err
	}

	filteredUsers := s.filterUsersByText(allUsers, search)
	total := int64(len(filteredUsers))

	return s.paginateUsers(filteredUsers, limit, offset), total, nil
}

func (s *UserService) filterUsersByText(users []*entities.User, search string) []*entities.User {
	if search == "" {
		return users
	}

	searchLower := strings.ToLower(search)
	var filtered []*entities.User

	for _, user := range users {
		if strings.Contains(strings.ToLower(user.Username), searchLower) ||
			strings.Contains(strings.ToLower(user.FirstName), searchLower) ||
			strings.Contains(strings.ToLower(user.LastName), searchLower) {
			filtered = append(filtered, user)
		}
	}

	return filtered
}

func (s *UserService) filterUsersByActiveStatus(users []*entities.User, isActive bool) []*entities.User {
	var filtered []*entities.User
	for _, user := range users {
		if user.IsActive == isActive {
			filtered = append(filtered, user)
		}
	}
	return filtered
}

func (s *UserService) paginateUsers(users []*entities.User, limit, offset int) []*entities.User {
	if offset >= len(users) {
		return []*entities.User{}
	}

	end := offset + limit
	if end > len(users) {
		end = len(users)
	}

	return users[offset:end]
}

func (s *UserService) toggleUserActiveStatus(ctx context.Context, id uint, isActive bool) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return entities.ErrUserNotFound
	}

	user.IsActive = isActive
	return s.userRepo.Update(ctx, user)
}
