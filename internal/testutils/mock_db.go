package testutils

import (
	"student-services-platform-backend/internal/db"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDB is a mock implementation of the gorm.DB interface
type MockDB struct {
	mock.Mock
}

// NewMockDB creates a new MockDB instance
func NewMockDB() *MockDB {
	return &MockDB{}
}

// Where mocks the gorm.DB Where method
func (m *MockDB) Where(query interface{}, args ...interface{}) *MockDB {
	m.Called(query, args)
	return m
}

// First mocks the gorm.DB First method
func (m *MockDB) First(dest interface{}, conds ...interface{}) *MockDB {
	m.Called(dest, conds)
	return m
}

// Create mocks the gorm.DB Create method
func (m *MockDB) Create(value interface{}) *MockDB {
	m.Called(value)
	return m
}

// Save mocks the gorm.DB Save method
func (m *MockDB) Save(value interface{}) *MockDB {
	m.Called(value)
	return m
}

// Model mocks the gorm.DB Model method
func (m *MockDB) Model(value interface{}) *MockDB {
	m.Called(value)
	return m
}

// Count mocks the gorm.DB Count method
func (m *MockDB) Count(count *int64) *MockDB {
	m.Called(count)
	return m
}

// Error mocks the gorm.DB Error method
func (m *MockDB) Error() error {
	args := m.Called()
	return args.Error(0)
}

// MockUserRepository is a mock implementation of the user repository functions
type MockUserRepository struct {
	mock.Mock
}

// NewMockUserRepository creates a new MockUserRepository instance
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{}
}

// GetUserByEmail mocks the GetUserByEmail function
func (m *MockUserRepository) GetUserByEmail(db *gorm.DB, email string) (*db.User, error) {
	args := m.Called(db, email)
	return args.Get(0).(*db.User), args.Error(1)
}

// CreateUser mocks the CreateUser function
func (m *MockUserRepository) CreateUser(db *gorm.DB, user *db.User) error {
	args := m.Called(db, user)
	return args.Error(0)
}

// GetUserByID mocks the GetUserByID function
func (m *MockUserRepository) GetUserByID(db *gorm.DB, id uint) (*db.User, error) {
	args := m.Called(db, id)
	return args.Get(0).(*db.User), args.Error(1)
}

// ExistsOtherUserWithEmail mocks the ExistsOtherUserWithEmail function
func (m *MockUserRepository) ExistsOtherUserWithEmail(db *gorm.DB, email string, excludeID uint) (bool, error) {
	args := m.Called(db, email, excludeID)
	return args.Bool(0), args.Error(1)
}

// UpdateUser mocks the UpdateUser function
func (m *MockUserRepository) UpdateUser(db *gorm.DB, user *db.User) error {
	args := m.Called(db, user)
	return args.Error(0)
}