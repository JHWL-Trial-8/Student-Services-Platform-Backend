package AuthController

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"student-services-platform-backend/app/services/auth"
	"student-services-platform-backend/internal/openapi"
	"student-services-platform-backend/internal/testutils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AuthControllerTestSuite struct {
	suite.Suite
	router       *gin.Engine
	mockAuthSvc  *MockAuthService
	userFactory  *testutils.UserFactory
}

// MockAuthService is a mock implementation of the auth service
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(req openapi.UserCreate) (*openapi.User, error) {
	args := m.Called(req)
	return args.Get(0).(*openapi.User), args.Error(1)
}

func (m *MockAuthService) Login(email, password string) (*openapi.AuthLoginPost200Response, error) {
	args := m.Called(email, password)
	return args.Get(0).(*openapi.AuthLoginPost200Response), args.Error(1)
}

func (suite *AuthControllerTestSuite) SetupTest() {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Create a test router
	suite.router = gin.New()
	suite.router.Use(gin.Recovery())
	
	// Create mock auth service
	suite.mockAuthSvc = &MockAuthService{}
	
	// Inject the mock service
	Svc = suite.mockAuthSvc
	
	// Setup routes
	suite.router.POST("/auth/login", AuthByPassword)
	suite.router.POST("/auth/register", RegisterByPassword)
}

func (suite *AuthControllerTestSuite) TestAuthByPassword_Success() {
	// Arrange
	loginReq := openapi.AuthLoginPostRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	
	expectedResponse := &openapi.AuthLoginPost200Response{
		AccessToken: "test-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}
	
	suite.mockAuthSvc.On("Login", loginReq.Email, loginReq.Password).Return(expectedResponse, nil)
	
	jsonValue, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	// Act
	suite.router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response openapi.AuthLoginPost200Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedResponse.AccessToken, response.AccessToken)
	assert.Equal(suite.T(), expectedResponse.TokenType, response.TokenType)
	assert.Equal(suite.T(), expectedResponse.ExpiresIn, response.ExpiresIn)
	
	suite.mockAuthSvc.AssertExpectations(suite.T())
}

func (suite *AuthControllerTestSuite) TestAuthByPassword_InvalidRequestBody() {
	// Arrange - create a request with invalid JSON
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	// Act
	suite.router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "请求参数错误", response["error"])
}

func (suite *AuthControllerTestSuite) TestAuthByPassword_UserNotFound() {
	// Arrange
	loginReq := openapi.AuthLoginPostRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}
	
	suite.mockAuthSvc.On("Login", loginReq.Email, loginReq.Password).
		Return(nil, &auth.ErrUserNotFound{Email: loginReq.Email})
	
	jsonValue, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	// Act
	suite.router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "用户不存在", response["error"])
	
	details := response["details"].(map[string]interface{})
	assert.Equal(suite.T(), loginReq.Email, details["email"])
	
	suite.mockAuthSvc.AssertExpectations(suite.T())
}

func (suite *AuthControllerTestSuite) TestAuthByPassword_InvalidPassword() {
	// Arrange
	loginReq := openapi.AuthLoginPostRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	
	suite.mockAuthSvc.On("Login", loginReq.Email, loginReq.Password).
		Return(nil, &auth.ErrInvalidPassword{Email: loginReq.Email})
	
	jsonValue, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	// Act
	suite.router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "密码错误", response["error"])
	
	details := response["details"].(map[string]interface{})
	assert.Equal(suite.T(), loginReq.Email, details["email"])
	
	suite.mockAuthSvc.AssertExpectations(suite.T())
}

func (suite *AuthControllerTestSuite) TestAuthByPassword_TokenGenerationError() {
	// Arrange
	loginReq := openapi.AuthLoginPostRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	
	suite.mockAuthSvc.On("Login", loginReq.Email, loginReq.Password).
		Return(nil, &auth.ErrGenerateToken{Message: "failed to generate token"})
	
	jsonValue, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	// Act
	suite.router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "生成令牌失败", response["error"])
	
	details := response["details"].(string)
	assert.Equal(suite.T(), "failed to generate token", details)
	
	suite.mockAuthSvc.AssertExpectations(suite.T())
}

func (suite *AuthControllerTestSuite) TestAuthByPassword_InternalError() {
	// Arrange
	loginReq := openapi.AuthLoginPostRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	
	suite.mockAuthSvc.On("Login", loginReq.Email, loginReq.Password).
		Return(nil, assert.AnError)
	
	jsonValue, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	// Act
	suite.router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "登录失败，请稍后再试", response["error"])
	
	suite.mockAuthSvc.AssertExpectations(suite.T())
}

func (suite *AuthControllerTestSuite) TestRegisterByPassword_Success() {
	// Arrange
	registerReq := openapi.UserCreate{
		Email:    "newuser@example.com",
		Name:     "New User",
		Password: "password123",
		Role:     openapi.RoleStudent,
		Phone:    "1234567890",
		Dept:     "Computer Science",
	}
	
	expectedUser := &openapi.User{
		Id:         1,
		Email:      registerReq.Email,
		Name:       registerReq.Name,
		Role:       registerReq.Role,
		Phone:      &registerReq.Phone,
		Dept:       &registerReq.Dept,
		IsActive:   true,
		AllowEmail: true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	suite.mockAuthSvc.On("Register", registerReq).Return(expectedUser, nil)
	
	jsonValue, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	// Act
	suite.router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
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
	
	suite.mockAuthSvc.AssertExpectations(suite.T())
}

func (suite *AuthControllerTestSuite) TestRegisterByPassword_InvalidRequestBody() {
	// Arrange - create a request with invalid JSON
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	// Act
	suite.router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "请求参数错误", response["error"])
}

func (suite *AuthControllerTestSuite) TestRegisterByPassword_EmailTaken() {
	// Arrange
	registerReq := openapi.UserCreate{
		Email:    "existing@example.com",
		Name:     "New User",
		Password: "password123",
		Role:     openapi.RoleStudent,
	}
	
	suite.mockAuthSvc.On("Register", registerReq).
		Return(nil, &auth.ErrEmailTaken{Email: registerReq.Email})
	
	jsonValue, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	// Act
	suite.router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "邮箱已被占用: "+registerReq.Email, response["error"])
	
	suite.mockAuthSvc.AssertExpectations(suite.T())
}

func TestAuthController(t *testing.T) {
	suite.Run(t, &AuthControllerTestSuite{})
}