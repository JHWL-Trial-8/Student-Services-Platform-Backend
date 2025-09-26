package UserController

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"student-services-platform-backend/app/services/user"
	"student-services-platform-backend/internal/openapi"
	"student-services-platform-backend/internal/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UserControllerTestSuite struct {
	suite.Suite
	router       *gin.Engine
	mockUserSvc  *MockUserService
}

// MockUserService is a mock implementation of the user service
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetUserInfomationByID(id string) (*openapi.User, error) {
	args := m.Called(id)
	return args.Get(0).(*openapi.User), args.Error(1)
}

func (m *MockUserService) UpdateUserInfomationByID(id uint, f user.UpdateFields) (*openapi.User, error) {
	args := m.Called(id, f)
	return args.Get(0).(*openapi.User), args.Error(1)
}

func (suite *UserControllerTestSuite) SetupTest() {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Create a test router
	suite.router = gin.New()
	suite.router.Use(gin.Recovery())
	
	// Create mock user service
	suite.mockUserSvc = &MockUserService{}
	
	// Inject the mock service
	Svc = suite.mockUserSvc
	
	// Setup routes
	suite.router.GET("/users/me", GetUserInform)
	suite.router.PUT("/users/me", UpdateMe)
}

func (suite *UserControllerTestSuite) TestGetUserInform_Success() {
	// Arrange
	userID := "123"
	expectedUser := &openapi.User{
		Id:         123,
		Email:      "test@example.com",
		Name:       "Test User",
		Role:       openapi.RoleStudent,
		Phone:      testutils.StringPtr("1234567890"),
		Dept:       testutils.StringPtr("Computer Science"),
		IsActive:   true,
		AllowEmail: true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	suite.mockUserSvc.On("GetUserInfomationByID", userID).Return(expectedUser, nil)
	
	req, _ := http.NewRequest("GET", "/users/me", nil)
	
	// Create a test gin context with user ID set
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("id", userID)
	
	// Act
	GetUserInform(c)
	
	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response openapi.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedUser.Id, response.Id)
	assert.Equal(suite.T(), expectedUser.Email, response.Email)
	assert.Equal(suite.T(), expectedUser.Name, response.Name)
	assert.Equal(suite.T(), expectedUser.Role, response.Role)
	assert.Equal(suite.T(), expectedUser.Phone, response.Phone)
	assert.Equal(suite.T(), expectedUser.Dept, response.Dept)
	assert.Equal(suite.T(), expectedUser.IsActive, response.IsActive)
	assert.Equal(suite.T(), expectedUser.AllowEmail, response.AllowEmail)
	
	suite.mockUserSvc.AssertExpectations(suite.T())
}

func (suite *UserControllerTestSuite) TestGetUserInform_MissingUserID() {
	// Arrange - create a request without user ID set
	req, _ := http.NewRequest("GET", "/users/me", nil)
	
	// Create a test gin context without user ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// Act
	GetUserInform(c)
	
	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "用户 ID 未找到", response["error"])
}

func (suite *UserControllerTestSuite) TestGetUserInform_ServiceError() {
	// Arrange
	userID := "123"
	suite.mockUserSvc.On("GetUserInfomationByID", userID).Return(nil, assert.AnError)
	
	req, _ := http.NewRequest("GET", "/users/me", nil)
	
	// Create a test gin context with user ID set
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("id", userID)
	
	// Act
	GetUserInform(c)
	
	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "获取用户失败", response["error"])
	assert.Equal(suite.T(), assert.AnError.Error(), response["details"])
	
	suite.mockUserSvc.AssertExpectations(suite.T())
}

func (suite *UserControllerTestSuite) TestUpdateMe_Success() {
	// Arrange
	userID := "123"
	userIDUint := uint(123)
	
	email := "updated@example.com"
	name := "Updated Name"
	phone := "9876543210"
	dept := "Updated Department"
	allowEmail := false
	
	updatePayload := map[string]interface{}{
		"email":       email,
		"name":        name,
		"phone":       phone,
		"dept":        dept,
		"allow_email": allowEmail,
	}
	
	expectedUser := &openapi.User{
		Id:         123,
		Email:      email,
		Name:       name,
		Role:       openapi.RoleStudent,
		Phone:      &phone,
		Dept:       &dept,
		IsActive:   true,
		AllowEmail: allowEmail,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	suite.mockUserSvc.On("UpdateUserInfomationByID", userIDUint, mock.AnythingOfType("user.UpdateFields")).Return(expectedUser, nil)
	
	jsonValue, _ := json.Marshal(updatePayload)
	req, _ := http.NewRequest("PUT", "/users/me", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	
	// Create a test gin context with user ID set
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("id", userID)
	
	// Act
	UpdateMe(c)
	
	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response openapi.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedUser.Id, response.Id)
	assert.Equal(suite.T(), expectedUser.Email, response.Email)
	assert.Equal(suite.T(), expectedUser.Name, response.Name)
	assert.Equal(suite.T(), expectedUser.Phone, response.Phone)
	assert.Equal(suite.T(), expectedUser.Dept, response.Dept)
	assert.Equal(suite.T(), expectedUser.AllowEmail, response.AllowEmail)
	
	suite.mockUserSvc.AssertExpectations(suite.T())
}

func (suite *UserControllerTestSuite) TestUpdateMe_MissingUserID() {
	// Arrange - create a request without user ID set
	updatePayload := map[string]interface{}{
		"email": "updated@example.com",
	}
	
	jsonValue, _ := json.Marshal(updatePayload)
	req, _ := http.NewRequest("PUT", "/users/me", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	
	// Create a test gin context without user ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// Act
	UpdateMe(c)
	
	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "用户 ID 未找到", response["error"])
}

func (suite *UserControllerTestSuite) TestUpdateMe_InvalidUserID() {
	// Arrange - create a request with invalid user ID
	userID := "invalid-id"
	
	updatePayload := map[string]interface{}{
		"email": "updated@example.com",
	}
	
	jsonValue, _ := json.Marshal(updatePayload)
	req, _ := http.NewRequest("PUT", "/users/me", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	
	// Create a test gin context with invalid user ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("id", userID)
	
	// Act
	UpdateMe(c)
	
	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "无效的用户 ID", response["error"])
}

func (suite *UserControllerTestSuite) TestUpdateMe_InvalidRequestBody() {
	// Arrange - create a request with invalid JSON
	userID := "123"
	
	req, _ := http.NewRequest("PUT", "/users/me", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	
	// Create a test gin context with user ID set
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("id", userID)
	
	// Act
	UpdateMe(c)
	
	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "请求参数错误", response["error"])
}

func (suite *UserControllerTestSuite) TestUpdateMe_MissingEmail() {
	// Arrange
	userID := "123"
	
	updatePayload := map[string]interface{}{
		"name": "Updated Name",
		// Missing email
	}
	
	jsonValue, _ := json.Marshal(updatePayload)
	req, _ := http.NewRequest("PUT", "/users/me", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	
	// Create a test gin context with user ID set
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("id", userID)
	
	// Act
	UpdateMe(c)
	
	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "email 为必填字段", response["error"])
}

func (suite *UserControllerTestSuite) TestUpdateMe_EmptyEmail() {
	// Arrange
	userID := "123"
	
	updatePayload := map[string]interface{}{
		"email": "", // Empty email
		"name":  "Updated Name",
	}
	
	jsonValue, _ := json.Marshal(updatePayload)
	req, _ := http.NewRequest("PUT", "/users/me", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	
	// Create a test gin context with user ID set
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("id", userID)
	
	// Act
	UpdateMe(c)
	
	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "email 为必填字段", response["error"])
}

func (suite *UserControllerTestSuite) TestUpdateMe_EmailTaken() {
	// Arrange
	userID := "123"
	userIDUint := uint(123)
	
	email := "taken@example.com"
	updatePayload := map[string]interface{}{
		"email": email,
		"name":  "Updated Name",
	}
	
	suite.mockUserSvc.On("UpdateUserInfomationByID", userIDUint, mock.AnythingOfType("user.UpdateFields")).
		Return(nil, &user.ErrEmailTaken{Email: email})
	
	jsonValue, _ := json.Marshal(updatePayload)
	req, _ := http.NewRequest("PUT", "/users/me", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	
	// Create a test gin context with user ID set
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("id", userID)
	
	// Act
	UpdateMe(c)
	
	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "邮箱已被占用", response["error"])
	
	details := response["details"].(map[string]interface{})
	assert.Equal(suite.T(), email, details["email"])
	
	suite.mockUserSvc.AssertExpectations(suite.T())
}

func (suite *UserControllerTestSuite) TestUpdateMe_ServiceError() {
	// Arrange
	userID := "123"
	userIDUint := uint(123)
	
	email := "updated@example.com"
	updatePayload := map[string]interface{}{
		"email": email,
		"name":  "Updated Name",
	}
	
	suite.mockUserSvc.On("UpdateUserInfomationByID", userIDUint, mock.AnythingOfType("user.UpdateFields")).
		Return(nil, assert.AnError)
	
	jsonValue, _ := json.Marshal(updatePayload)
	req, _ := http.NewRequest("PUT", "/users/me", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	
	// Create a test gin context with user ID set
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("id", userID)
	
	// Act
	UpdateMe(c)
	
	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "更新失败", response["error"])
	assert.Equal(suite.T(), assert.AnError.Error(), response["details"])
	
	suite.mockUserSvc.AssertExpectations(suite.T())
}

func TestUserController(t *testing.T) {
	suite.Run(t, &UserControllerTestSuite{})
}