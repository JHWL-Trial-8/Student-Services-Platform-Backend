package adminuser

import (
	"fmt"
	"strconv"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	dbpkg "student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/openapi"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ListUsers retrieves users with pagination and optional role filtering
func (s *Service) ListUsers(page, pageSize int, role *openapi.Role) (*openapi.PagedUsers, error) {
	var dbRole *dbpkg.Role
	if role != nil {
		r := dbpkg.Role(*role)
		dbRole = &r
	}

	users, total, err := dbpkg.ListUsers(s.db, page, pageSize, dbRole)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	items := make([]openapi.User, len(users))
	for i, user := range users {
		items[i] = openapi.User{
			Id:         int32(user.ID),
			Email:      user.Email,
			Name:       user.Name,
			Role:       openapi.Role(user.Role),
			Phone:      user.Phone,
			Dept:       user.Dept,
			IsActive:   user.IsActive,
			AllowEmail: user.AllowEmail,
			CreatedAt:  user.CreatedAt,
			UpdatedAt:  user.UpdatedAt,
		}
	}

	return &openapi.PagedUsers{
		Items:    items,
		Page:     int32(page),
		PageSize: int32(pageSize),
		Total:    int32(total),
	}, nil
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(id string) (*openapi.User, error) {
	idUint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := dbpkg.GetUserByID(s.db, uint(idUint))
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &openapi.User{
		Id:         int32(user.ID),
		Email:      user.Email,
		Name:       user.Name,
		Role:       openapi.Role(user.Role),
		Phone:      user.Phone,
		Dept:       user.Dept,
		IsActive:   user.IsActive,
		AllowEmail: user.AllowEmail,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}, nil
}

// CreateUser creates a new user with password hashing
func (s *Service) CreateUser(req openapi.UserCreate) (*openapi.User, error) {
	// Check if email is already taken
	if _, err := dbpkg.GetUserByEmail(s.db, req.Email); err == nil {
		return nil, &ErrEmailTaken{Email: req.Email}
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &dbpkg.User{
		Email:        req.Email,
		Name:         req.Name,
		Role:         dbpkg.Role(req.Role),
		Phone:        nilIfEmpty(req.Phone),
		Dept:         nilIfEmpty(req.Dept),
		IsActive:     req.IsActive,
		AllowEmail:   req.AllowEmail,
		PasswordHash: string(hash),
	}

	if err := dbpkg.CreateUser(s.db, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &openapi.User{
		Id:         int32(user.ID),
		Email:      user.Email,
		Name:       user.Name,
		Role:       openapi.Role(user.Role),
		Phone:      user.Phone,
		Dept:       user.Dept,
		IsActive:   user.IsActive,
		AllowEmail: user.AllowEmail,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}, nil
}

// UpdateUser updates a user by ID
func (s *Service) UpdateUser(id string, req openapi.UserAdminUpdate) (*openapi.User, error) {
	idUint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Get user
	user, err := dbpkg.GetUserByID(s.db, uint(idUint))
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check email uniqueness if email is being changed
	if req.Email != user.Email {
		if taken, err := dbpkg.ExistsOtherUserWithEmail(s.db, req.Email, user.ID); err != nil {
			return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
		} else if taken {
			return nil, &ErrEmailTaken{Email: req.Email}
		}
		user.Email = req.Email
	}

	// Update other fields
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Phone != "" {
		user.Phone = &req.Phone
	} else if req.Phone == "" {
		user.Phone = nil
	}
	if req.Dept != "" {
		user.Dept = &req.Dept
	} else if req.Dept == "" {
		user.Dept = nil
	}
	if req.Role != "" {
		user.Role = dbpkg.Role(req.Role)
	}
	// IsActive and AllowEmail are always included in the request
	user.IsActive = req.IsActive
	user.AllowEmail = req.AllowEmail

	if err := dbpkg.UpdateUser(s.db, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &openapi.User{
		Id:         int32(user.ID),
		Email:      user.Email,
		Name:       user.Name,
		Role:       openapi.Role(user.Role),
		Phone:      user.Phone,
		Dept:       user.Dept,
		IsActive:   user.IsActive,
		AllowEmail: user.AllowEmail,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}, nil
}

// DeleteUser deletes a user by ID
func (s *Service) DeleteUser(id string) error {
	idUint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	if err := dbpkg.DeleteUser(s.db, uint(idUint)); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// Helper function to return nil for empty strings
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ErrEmailTaken when email is already taken
type ErrEmailTaken struct{ Email string }

func (e *ErrEmailTaken) Error() string { return fmt.Sprintf("email already taken: %s", e.Email) }