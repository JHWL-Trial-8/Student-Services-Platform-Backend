package midwares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type JWTMiddlewareTestSuite struct {
	suite.Suite
	secretKey string
	router    *gin.Engine
}

func (suite *JWTMiddlewareTestSuite) SetupTest() {
	suite.secretKey = "test-secret-key"
	
	// Set gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Create a test router with JWT middleware
	suite.router = gin.New()
	suite.router.Use(JWTAuthMidware(suite.secretKey))
	
	// Add a test endpoint
	suite.router.GET("/test", func(c *gin.Context) {
		userID := c.GetString("id")
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})
}

func (suite *JWTMiddlewareTestSuite) generateTestToken(userID string, expiration time.Time) string {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(expiration),
		Issuer:    "test-issuer",
		Audience:  []string{"test-audience"},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(suite.secretKey))
	require.NoError(suite.T(), err)
	return tokenString
}

func (suite *JWTMiddlewareTestSuite) TestJWTAuthMidware_ValidToken() {
	// Arrange
	userID := "123"
	validToken := suite.generateTestToken(userID, time.Now().Add(time.Hour))
	
	// Create a test request
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	
	// Create a response recorder
	w := httptest.NewRecorder()
	
	// Act
	suite.router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	// Verify the response contains the user ID
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), userID, response["user_id"])
}

func (suite *JWTMiddlewareTestSuite) TestJWTAuthMidware_MissingAuthorizationHeader() {
	// Arrange - create a request without Authorization header
	req, _ := http.NewRequest("GET", "/test", nil)
	
	// Create a response recorder
	w := httptest.NewRecorder()
	
	// Act
	suite.router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	
	// Verify the error message
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(http.StatusUnauthorized), response["code"])
	assert.Equal(suite.T(), "未识别到Token，请登录后访问", response["error"])
}

func (suite *JWTMiddlewareTestSuite) TestJWTAuthMidware_InvalidAuthorizationFormat() {
	// Arrange - create a request with invalid Authorization format
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat token")
	
	// Create a response recorder
	w := httptest.NewRecorder()
	
	// Act
	suite.router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	
	// Verify the error message
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(http.StatusUnauthorized), response["code"])
	assert.Equal(suite.T(), "Token 格式错误，需为 Bearer <token>", response["error"])
}

func (suite *JWTMiddlewareTestSuite) TestJWTAuthMidware_ExpiredToken() {
	// Arrange
	userID := "123"
	expiredToken := suite.generateTestToken(userID, time.Now().Add(-time.Hour)) // Token expired 1 hour ago
	
	// Create a test request
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	
	// Create a response recorder
	w := httptest.NewRecorder()
	
	// Act
	suite.router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	
	// Verify the error message
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(http.StatusUnauthorized), response["code"])
	assert.Equal(suite.T(), "Token 已过期，请重新登录", response["error"])
}

func (suite *JWTMiddlewareTestSuite) TestJWTAuthMidware_InvalidSignature() {
	// Arrange - create a token with a different secret key
	userID := "123"
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		Issuer:    "test-issuer",
		Audience:  []string{"test-audience"},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Sign with a different secret key
	tokenString, err := token.SignedString([]byte("different-secret-key"))
	require.NoError(suite.T(), err)
	
	// Create a test request
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	
	// Create a response recorder
	w := httptest.NewRecorder()
	
	// Act
	suite.router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	
	// Verify the error message
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(http.StatusUnauthorized), response["code"])
	assert.Equal(suite.T(), "Token 无效", response["error"])
}

func (suite *JWTMiddlewareTestSuite) TestJWTAuthMidware_EmptySubject() {
	// Arrange - create a token with empty subject
	claims := jwt.RegisteredClaims{
		Subject:   "", // Empty subject
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		Issuer:    "test-issuer",
		Audience:  []string{"test-audience"},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(suite.secretKey))
	require.NoError(suite.T(), err)
	
	// Create a test request
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	
	// Create a response recorder
	w := httptest.NewRecorder()
	
	// Act
	suite.router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	
	// Verify the error message
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(http.StatusUnauthorized), response["code"])
	assert.Equal(suite.T(), "Token 中无用户信息", response["error"])
}

func (suite *JWTMiddlewareTestSuite) TestJWTAuthMidware_InvalidSigningMethod() {
	// Arrange - create a token with an invalid signing method
	userID := "123"
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		Issuer:    "test-issuer",
		Audience:  []string{"test-audience"},
	}
	
	// Use a different signing method (HS512 instead of HS256)
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString([]byte(suite.secretKey))
	require.NoError(suite.T(), err)
	
	// Create a test request
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	
	// Create a response recorder
	w := httptest.NewRecorder()
	
	// Act
	suite.router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	
	// Verify the error message
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(http.StatusUnauthorized), response["code"])
	assert.Equal(suite.T(), "无效的签名算法，仅支持 HS256", response["error"])
}

func (suite *JWTMiddlewareTestSuite) TestJWTAuthMidware_MalformedToken() {
	// Arrange - create a request with a malformed token
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer this.is.not.a.valid.token")
	
	// Create a response recorder
	w := httptest.NewRecorder()
	
	// Act
	suite.router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	
	// Verify the error message
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(http.StatusUnauthorized), response["code"])
	assert.Equal(suite.T(), "Token 格式错误", response["error"])
}

func TestJWTMiddleware(t *testing.T) {
	suite.Run(t, &JWTMiddlewareTestSuite{})
}