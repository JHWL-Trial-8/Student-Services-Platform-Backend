package testutils

import (
	"testing"
	"time"

	"student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/openapi"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// UserFactory creates test users with default or custom values
type UserFactory struct {
	db *gorm.DB
	t  *testing.T
}

// NewUserFactory creates a new UserFactory instance
func NewUserFactory(db *gorm.DB, t *testing.T) *UserFactory {
	return &UserFactory{
		db: db,
		t:  t,
	}
}

// CreateDefaultUser creates a user with default values
func (f *UserFactory) CreateDefaultUser() *db.User {
	user := &db.User{
		Email:        "test@example.com",
		Name:         "Test User",
		Role:         db.RoleStudent,
		Phone:        nil,
		Dept:         nil,
		IsActive:     true,
		AllowEmail:   true,
		PasswordHash: "$2a$10$EixZaYVK1fsbw1ZfbX3OXePaWxn96p36WQoeG6Lruj3vjPkU2Cslq", // password: "secret"
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := f.db.Create(user).Error
	require.NoError(f.t, err, "Failed to create default test user")
	return user
}

// CreateAdminUser creates an admin user with default values
func (f *UserFactory) CreateAdminUser() *db.User {
	user := &db.User{
		Email:        "admin@example.com",
		Name:         "Admin User",
		Role:         db.RoleAdmin,
		Phone:        nil,
		Dept:         nil,
		IsActive:     true,
		AllowEmail:   true,
		PasswordHash: "$2a$10$EixZaYVK1fsbw1ZfbX3OXePaWxn96p36WQoeG6Lruj3vjPkU2Cslq", // password: "secret"
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := f.db.Create(user).Error
	require.NoError(f.t, err, "Failed to create admin test user")
	return user
}

// CreateCustomUser creates a user with custom values
func (f *UserFactory) CreateCustomUser(email, name string, role db.Role, password string) *db.User {
	user := &db.User{
		Email:        email,
		Name:         name,
		Role:         role,
		Phone:        nil,
		Dept:         nil,
		IsActive:     true,
		AllowEmail:   true,
		PasswordHash: password,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := f.db.Create(user).Error
	require.NoError(f.t, err, "Failed to create custom test user")
	return user
}

// ToOpenAPIUser converts a db.User to openapi.User
func ToOpenAPIUser(user *db.User) *openapi.User {
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
	}
}

// StringPtr returns a pointer to the provided string
func StringPtr(s string) *string {
	return &s
}

// BoolPtr returns a pointer to the provided bool
func BoolPtr(b bool) *bool {
	return &b
}

// SetupTestDatabase creates and migrates the test database
func SetupTestDatabase(t *testing.T) *gorm.DB {
	// Import here to avoid circular dependency
	cfg := &struct {
		Database struct {
			Driver string `mapstructure:"driver"`
			DSN    string `mapstructure:"dsn"`
		}
	}{
		Database: struct {
			Driver string `mapstructure:"driver"`
			DSN    string `mapstructure:"dsn"`
		}{
			Driver: "sqlite",
			DSN:    "file::memory:?cache=shared",
		},
	}

	// Connect to in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(cfg.Database.DSN), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to test database")

	// Auto migrate all models
	err = db.AutoMigrate(
		&db.User{},
		&db.Ticket{},
		&db.TicketMessage{},
		&db.Image{},
		&db.TicketImage{},
		&db.Rating{},
		&db.AuditLog{},
		&db.SpamFlag{},
		&db.CannedReply{},
	)
	require.NoError(t, err, "Failed to migrate test database")

	// Clean up function to be called after test
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err != nil {
			t.Logf("Error getting underlying sql.DB: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			t.Logf("Error closing test database: %v", err)
		}
	})

	return db
}