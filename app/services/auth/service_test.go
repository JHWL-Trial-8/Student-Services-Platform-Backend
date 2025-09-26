package auth

import (
	"testing"
	"time"

	"student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/openapi"
	"student-services-platform-backend/internal/testutils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type AuthServiceTestSuite struct {
	testutils.TestSuite
	service *Service
}

func (suite *AuthServiceTestSuite) SetupTest() {
	suite.TestSuite.SetupTest()
	
	// Create auth service with test configuration
	suite.service = NewService(suite.DB, &JWTConfig{
		SecretKey:      "test-secret-key",
		AccessTokenExp: time.Hour,
		Issuer:         "test-issuer",
		Audience:       "test-audience",
	})
}

func (suite *AuthServiceTestSuite) TestRegister_Success() {
	// Arrange
	userReq := openapi.UserCreate{
		Email:    "newuser@example.com",
		Name:     "New User",
		Password: "password123",
		Role:     openapi.RoleStudent,
		Phone:    "1234567890",
		Dept:     "Computer Science",
	}

	// Act
	user, err := suite.service.Register(userReq)

	// Assert
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), userReq.Email, user.Email)
	assert.Equal(suite.T(), userReq.Name, user.Name)
	assert.Equal(suite.T(), userReq.Role, user.Role)
	assert.Equal(suite.T(), userReq.Phone, user.Phone)
	assert.Equal(suite.T(), userReq.Dept, user.Dept)
	assert.True(suite.T(), user.IsActive)
	assert.True(suite.T(), user.AllowEmail)
	assert.NotZero(suite.T(), user.CreatedAt)
	assert.NotZero(suite.T(), user.UpdatedAt)

	// Verify user was actually created in the database
	var dbUser db.User
	err = suite.DB.Where("email = ?", userReq.Email).First(&dbUser).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), userReq.Name, dbUser.Name)
	assert.Equal(suite.T(), db.Role(userReq.Role), dbUser.Role)
}

func (suite *AuthServiceTestSuite) TestRegister_EmailAlreadyTaken() {
	// Arrange - create a user first
	existingUser := suite.CreateTestUser("existing@example.com", "Existing User", db.RoleStudent)
	
	userReq := openapi.UserCreate{
		Email:    existingUser.Email, // Same email as existing user
		Name:     "New User",
		Password: "password123",
		Role:     openapi.RoleStudent,
	}

	// Act
	user, err := suite.service.Register(userReq)

	// Assert
	assert.Nil(suite.T(), user)
	assert.Error(suite.T(), err)
	assert.IsType(suite.T(), &ErrEmailTaken{}, err)
	assert.Contains(suite.T(), err.Error(), existingUser.Email)
}

func (suite *AuthServiceTestSuite) TestRegister_DatabaseError() {
	// This test would require mocking the database to return an error
	// For now, we'll skip it as it would require more complex setup
	suite.T().Skip("Skipping database error test - would require more complex mock setup")
}

func (suite *AuthServiceTestSuite) TestLogin_Success() {
	// Arrange - create a user with known password
	user := suite.CreateTestUser("login@example.com", "Login User", db.RoleStudent)
	
	// Act
	response, err := suite.service.Login(user.Email, "secret") // "secret" is the password for our test hash

	// Assert
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.NotEmpty(suite.T(), response.AccessToken)
	assert.Equal(suite.T(), "Bearer", response.TokenType)
	assert.Equal(suite.T(), int32(time.Hour.Seconds()), response.ExpiresIn)

	// Verify the JWT token is valid
	token, err := jwt.ParseWithClaims(response.AccessToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key"), nil
	})
	require.NoError(suite.T(), err)
	assert.True(suite.T(), token.Valid)

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	require.True(suite.T(), ok)
	assert.Equal(suite.T(), "test-issuer", claims.Issuer)
	assert.Equal(suite.T(), "test-audience", claims.Audience[0])
	assert.Equal(suite.T(), string(user.ID), claims.Subject)
}

func (suite *AuthServiceTestSuite) TestLogin_UserNotFound() {
	// Act
	response, err := suite.service.Login("nonexistent@example.com", "password")

	// Assert
	assert.Nil(suite.T(), response)
	assert.Error(suite.T(), err)
	assert.IsType(suite.T(), &ErrUserNotFound{}, err)
	assert.Contains(suite.T(), err.Error(), "nonexistent@example.com")
}

func (suite *AuthServiceTestSuite) TestLogin_InvalidPassword() {
	// Arrange - create a user
	user := suite.CreateTestUser("invalidpass@example.com", "Invalid Pass User", db.RoleStudent)
	
	// Act
	response, err := suite.service.Login(user.Email, "wrongpassword")

	// Assert
	assert.Nil(suite.T(), response)
	assert.Error(suite.T(), err)
	assert.IsType(suite.T(), &ErrInvalidPassword{}, err)
	assert.Contains(suite.T(), err.Error(), user.Email)
}

func (suite *AuthServiceTestSuite) TestGenerateAccessToken_Success() {
	// Arrange - create a user
	user := suite.CreateTestUser("token@example.com", "Token User", db.RoleStudent)

	// Act
	tokenResp, err := suite.service.generateAccessToken(user)

	// Assert
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), tokenResp)
	assert.NotEmpty(suite.T(), tokenResp.AccessToken)
	assert.Equal(suite.T(), time.Hour, tokenResp.ExpiresIn)

	// Verify the JWT token is valid
	token, err := jwt.ParseWithClaims(tokenResp.AccessToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key"), nil
	})
	require.NoError(suite.T(), err)
	assert.True(suite.T(), token.Valid)

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	require.True(suite.T(), ok)
	assert.Equal(suite.T(), "test-issuer", claims.Issuer)
	assert.Equal(suite.T(), "test-audience", claims.Audience[0])
	assert.Equal(suite.T(), string(user.ID), claims.Subject)
}

func (suite *AuthServiceTestSuite) TestGenerateAccessToken_EmptySecretKey() {
	// Arrange - create a service with empty secret key
	service := NewService(suite.DB, &JWTConfig{
		SecretKey:      "",
		AccessTokenExp: time.Hour,
		Issuer:         "test-issuer",
		Audience:       "test-audience",
	})
	
	user := suite.CreateTestUser("emptykey@example.com", "Empty Key User", db.RoleStudent)

	// Act
	tokenResp, err := service.generateAccessToken(user)

	// Assert
	assert.Nil(suite.T(), tokenResp)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "JWT 密钥未配置")
}

func (suite *AuthServiceTestSuite) TestGenerateAccessToken_InvalidExpiration() {
	// Arrange - create a service with invalid expiration
	service := NewService(suite.DB, &JWTConfig{
		SecretKey:      "test-secret-key",
		AccessTokenExp: 0, // Invalid expiration
		Issuer:         "test-issuer",
		Audience:       "test-audience",
	})
	
	user := suite.CreateTestUser("invalidexp@example.com", "Invalid Exp User", db.RoleStudent)

	// Act
	tokenResp, err := service.generateAccessToken(user)

	// Assert
	assert.Nil(suite.T(), tokenResp)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "JWT 访问令牌有效期无效")
}

func TestAuthService(t *testing.T) {
	testutils.RunTestSuite(t, &AuthServiceTestSuite{})
}