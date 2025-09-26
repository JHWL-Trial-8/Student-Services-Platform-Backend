package testutils

import (
	"testing"

	"student-services-platform-backend/internal/db"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// TestSuite is a base test suite that provides common setup and teardown
type TestSuite struct {
	suite.Suite
	DB        *gorm.DB
	UserRepo  *MockUserRepository
	UserFactory *UserFactory
}

// SetupTest is called before each test
func (s *TestSuite) SetupTest() {
	// Setup test database
	s.DB = SetupTestDatabase(s.T())
	
	// Setup mock repository
	s.UserRepo = NewMockUserRepository()
	
	// Setup user factory
	s.UserFactory = NewUserFactory(s.DB, s.T())
}

// TearDownTest is called after each test
func (s *TestSuite) TearDownTest() {
	// Clean up database
	if s.DB != nil {
		sqlDB, err := s.DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

// CreateTestUser creates a test user with the given parameters
func (s *TestSuite) CreateTestUser(email, name string, role db.Role) *db.User {
	return s.UserFactory.CreateCustomUser(email, name, role, "$2a$10$EixZaYVK1fsbw1ZfbX3OXePaWxn96p36WQoeG6Lruj3vjPkU2Cslq")
}

// CreateDefaultTestUser creates a test user with default values
func (s *TestSuite) CreateDefaultTestUser() *db.User {
	return s.UserFactory.CreateDefaultUser()
}

// CreateAdminTestUser creates an admin test user with default values
func (s *TestSuite) CreateAdminTestUser() *db.User {
	return s.UserFactory.CreateAdminUser()
}

// RunTestSuite runs the given test suite
func RunTestSuite(t *testing.T, s suite.TestingSuite) {
	suite.Run(t, s)
}