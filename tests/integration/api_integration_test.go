package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"student-services-platform-backend/app/controllers/AuthController"
	"student-services-platform-backend/app/controllers/UserController"
	"student-services-platform-backend/app/router"
	"student-services-platform-backend/app/services/auth"
	usersvc "student-services-platform-backend/app/services/user"
	"student-services-platform-backend/internal/config"
	"student-services-platform-backend/internal/db"
	httpmw "student-services-platform-backend/internal/http"
	"student-services-platform-backend/internal/openapi"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type APIIntegrationTestSuite struct {
	suite.Suite
	router *gin.Engine
	db     *gorm.DB
}

func (suite *APIIntegrationTestSuite) SetupSuite() {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Setup test database
	suite.db = config.TestDB(suite.T())
	
	// Auto migrate all models
	err := suite.db.AutoMigrate(
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
	require.NoError(suite.T(), err, "Failed to migrate test database")
	
	// Setup the application
	cfg := config.TestConfig(suite.T())
	
	// Build auth service
	accessExp, _ := time.ParseDuration(cfg.JWT.AccessTokenExp)
	authSvc := auth.NewService(suite.db, &auth.JWTConfig{
		SecretKey:      cfg.JWT.SecretKey,
		AccessTokenExp: accessExp,
		Issuer:         cfg.JWT.Issuer,
		Audience:       cfg.JWT.Audience,
	})
	
	// Inject services
	AuthController.Svc = authSvc
	UserController.Svc = usersvc.NewService(suite.db)
	
	// Setup router
	suite.router = gin.New()
	suite.router.Use(gin.Logger(), gin.Recovery(), httpmw.CORS(cfg.CORS))
	
	api := suite.router.Group("/api/v1")
	{
		api.GET("/healthz", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true, "ts": time.Now().UTC().Format(time.RFC3339)})
		})
		router.Init(api, cfg)
	}
}

func (suite *APIIntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		sqlDB, err := suite.db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

func (suite *APIIntegrationTestSuite) SetupTest() {
	// Clean up database before each test
	suite.db.Exec("DELETE FROM users")
}

func (suite *APIIntegrationTestSuite) TestHealthCheck() {
	// Arrange
	req, _ := http.NewRequest("GET", "/api/v1/healthz", nil)
	
	// Create a response recorder
	w := httptest.NewRecorder()
	
	// Act
	suite.router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response["ok"].(bool))
	assert.NotEmpty(suite.T(), response["ts"])
}

func (suite *APIIntegrationTestSuite) TestRegisterAndLoginFlow() {
	// Arrange - register request
	registerReq := openapi.UserCreate{
		Email:    "integration@example.com",
		Name:     "Integration User",
		Password: "password123",
		Role:     openapi.RoleStudent,
		Phone:    "1234567890",
		Dept:     "Computer Science",
	}
	
	jsonValue, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	// Act - register
	suite.router.ServeHTTP(w, req)
	
	// Assert - register response
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var registerResponse openapi.User
	err := json.Unmarshal(w.Body.Bytes(), &registerResponse)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), registerReq.Email, registerResponse.Email)
	assert.Equal(suite.T(), registerReq.Name, registerResponse.Name)
	assert.Equal(suite.T(), registerReq.Role, registerResponse.Role)
	assert.Equal(suite.T(), &registerReq.Phone, registerResponse.Phone)
	assert.Equal(suite.T(), &registerReq.Dept, registerResponse.Dept)
	assert.True(suite.T(), registerResponse.IsActive)
	assert.True(suite.T(), registerResponse.AllowEmail)
	assert.NotZero(suite.T(), registerResponse.Id)
	
	// Arrange - login request
	loginReq := openapi.AuthLoginPostRequest{
		Email:    registerReq.Email,
		Password: registerReq.Password,
	}
	
	jsonValue, _ = json.Marshal(loginReq)
	req, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	
	w = httptest.NewRecorder()
	
	// Act - login
	suite.router.ServeHTTP(w, req)
	
	// Assert - login response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var loginResponse openapi.AuthLoginPost200Response
	err = json.Unmarshal(w.Body.Bytes(), &loginResponse)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), loginResponse.AccessToken)
	assert.Equal(suite.T(), "Bearer", loginResponse.TokenType)
	assert.Equal(suite.T(), int32(3600), loginResponse.ExpiresIn)
	
	// Arrange - get user info request
	req, _ = http.NewRequest("GET", "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)
	
	w = httptest.NewRecorder()
	
	// Act - get user info
	suite.router.ServeHTTP(w, req)
	
	// Assert - get user info response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var userInfo openapi.User
	err = json.Unmarshal(w.Body.Bytes(), &userInfo)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), registerResponse.Id, userInfo.Id)
	assert.Equal(suite.T(), registerResponse.Email, userInfo.Email)
	assert.Equal(suite.T(), registerResponse.Name, userInfo.Name)
	assert.Equal(suite.T(), registerResponse.Role, userInfo.Role)
	assert.Equal(suite.T(), registerResponse.Phone, userInfo.Phone)
	assert.Equal(suite.T(), registerResponse.Dept, userInfo.Dept)
	assert.Equal(suite.T(), registerResponse.IsActive, userInfo.IsActive)
	assert.Equal(suite.T(), registerResponse.AllowEmail, userInfo.AllowEmail)
}

func (suite *APIIntegrationTestSuite) TestUpdateUserFlow() {
	// First, register and login to get a token
	token := suite.registerAndLogin("update@example.com", "Update User", "password123")
	
	// Arrange - update user request
	updateReq := map[string]interface{}{
		"email":       "updated@example.com",
		"name":        "Updated Name",
		"phone":       "9876543210",
		"dept":        "Updated Department",
		"allow_email": false,
	}
	
	jsonValue, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/api/v1/users/me", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	// Act - update user
	suite.router.ServeHTTP(w, req)
	
	// Assert - update response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var updateResponse openapi.User
	err := json.Unmarshal(w.Body.Bytes(), &updateResponse)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), updateReq["email"], updateResponse.Email)
	assert.Equal(suite.T(), updateReq["name"], updateResponse.Name)
	assert.Equal(suite.T(), updateReq["phone"], *updateResponse.Phone)
	assert.Equal(suite.T(), updateReq["dept"], *updateResponse.Dept)
	assert.Equal(suite.T(), updateReq["allow_email"], updateResponse.AllowEmail)
	
	// Verify the update by getting user info again
	req, _ = http.NewRequest("GET", "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	w = httptest.NewRecorder()
	
	// Act - get updated user info
	suite.router.ServeHTTP(w, req)
	
	// Assert - get updated user info response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var userInfo openapi.User
	err = json.Unmarshal(w.Body.Bytes(), &userInfo)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), updateResponse.Id, userInfo.Id)
	assert.Equal(suite.T(), updateResponse.Email, userInfo.Email)
	assert.Equal(suite.T(), updateResponse.Name, userInfo.Name)
	assert.Equal(suite.T(), updateResponse.Phone, userInfo.Phone)
	assert.Equal(suite.T(), updateResponse.Dept, userInfo.Dept)
	assert.Equal(suite.T(), updateResponse.AllowEmail, userInfo.AllowEmail)
}

func (suite *APIIntegrationTestSuite) TestInvalidToken() {
	// Arrange - request with invalid token
	req, _ := http.NewRequest("GET", "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	
	w := httptest.NewRecorder()
	
	// Act
	suite.router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(http.StatusUnauthorized), response["code"])
	assert.Equal(suite.T(), "Token 无效", response["error"])
}

func (suite *APIIntegrationTestSuite) TestMissingToken() {
	// Arrange - request without token
	req, _ := http.NewRequest("GET", "/api/v1/users/me", nil)
	
	w := httptest.NewRecorder()
	
	// Act
	suite.router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(http.StatusUnauthorized), response["code"])
	assert.Equal(suite.T(), "未识别到Token，请登录后访问", response["error"])
}

func (suite *APIIntegrationTestSuite) TestDuplicateEmailRegistration() {
	// Arrange - register first user
	registerReq := openapi.UserCreate{
		Email:    "duplicate@example.com",
		Name:     "First User",
		Password: "password123",
		Role:     openapi.RoleStudent,
	}
	
	jsonValue, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	// Act - register first user
	suite.router.ServeHTTP(w, req)
	
	// Assert - first registration should succeed
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	// Arrange - try to register another user with the same email
	registerReq.Name = "Second User"
	
	jsonValue, _ = json.Marshal(registerReq)
	req, _ = http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	
	w = httptest.NewRecorder()
	
	// Act - try to register second user
	suite.router.ServeHTTP(w, req)
	
	// Assert - second registration should fail
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "邮箱已被占用: "+registerReq.Email, response["error"])
}

// registerAndLogin is a helper method that registers a user and returns an auth token
func (suite *APIIntegrationTestSuite) registerAndLogin(email, name, password string) string {
	// Register user
	registerReq := openapi.UserCreate{
		Email:    email,
		Name:     name,
		Password: password,
		Role:     openapi.RoleStudent,
	}
	
	jsonValue, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// Login user
	loginReq := openapi.AuthLoginPostRequest{
		Email:    email,
		Password: password,
	}
	
	jsonValue, _ = json.Marshal(loginReq)
	req, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	var loginResponse openapi.AuthLoginPost200Response
	err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
	require.NoError(suite.T(), err)
	
	return loginResponse.AccessToken
}

func TestAPIIntegration(t *testing.T) {
	suite.Run(t, &APIIntegrationTestSuite{})
}