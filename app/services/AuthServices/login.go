package AuthService

import (
	"errors"
	"fmt"
	"student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/openapi"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db  *gorm.DB
	cfg *JWTConfig
}

type JWTConfig struct {
	SecretKey      string
	AccessTokenExp time.Duration
	Issuer         string
	Audience       string
}

type ErrUserNotFound struct {
	Email string
}

func (e *ErrUserNotFound) Error() string {
	return fmt.Sprintf("用户不存在: %s", e.Email)
}

type ErrInvalidPassword struct {
	Email string
}

func (e *ErrInvalidPassword) Error() string {
	return fmt.Sprintf("密码错误: %s", e.Email)
}

type ErrGenerateToken struct {
	Message string
}

func (e *ErrGenerateToken) Error() string {
	return fmt.Sprintf("生成令牌失败: %s", e.Message)
}

func NewAuthService(db *gorm.DB, cfg *JWTConfig) *AuthService {
	return &AuthService{db: db, cfg: cfg}
}

func (as *AuthService) Login(email, password string) (*openapi.AuthLoginPost200Response, error) {
	dbUser := &db.User{}
	if err := as.db.Where("email = ?", email).First(dbUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &ErrUserNotFound{Email: email}
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte(password)); err != nil {
		return nil, &ErrInvalidPassword{Email: email}
	}

	tokenResp, err := generateAccessToken(dbUser, as.cfg)
	if err != nil {
		return nil, &ErrGenerateToken{Message: err.Error()}
	}

	return &openapi.AuthLoginPost200Response{
		AccessToken: tokenResp.AccessToken,
		TokenType:   "bearer",
		ExpiresIn:   int32(tokenResp.ExpiresIn.Seconds()),
	}, nil
}

func generateAccessToken(dbUser *db.User, cfg *JWTConfig) (*struct {
	AccessToken string        `json:"access_token"`
	ExpiresIn   time.Duration `json:"expires_in"`
}, error) {
	if cfg.SecretKey == "" {
		return nil, errors.New("JWT 密钥未配置")
	}
	if cfg.AccessTokenExp <= 0 {
		return nil, errors.New("JWT 访问令牌有效期无效")
	}

	apiUser := openapi.User{
		Id:         int32(dbUser.ID),
		Email:      dbUser.Email,
		Name:       dbUser.Name,
		Role:       openapi.Role(dbUser.Role),
		Phone:      dbUser.Phone,
		Dept:       dbUser.Dept,
		IsActive:   dbUser.IsActive,
		AllowEmail: dbUser.AllowEmail,
		CreatedAt:  dbUser.CreatedAt,
		UpdatedAt:  dbUser.UpdatedAt,
	}

	accessTokenClaims := jwt.RegisteredClaims{
		Subject:   string(apiUser.Id),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.AccessTokenExp)),
		Issuer:    cfg.Issuer,
		Audience:  []string{cfg.Audience},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	tokenString, err := token.SignedString([]byte(cfg.SecretKey))
	if err != nil {
		return nil, fmt.Errorf("JWT 签名失败: %w", err)
	}

	return &struct {
		AccessToken string        `json:"access_token"`
		ExpiresIn   time.Duration `json:"expires_in"`
	}{
		AccessToken: tokenString,
		ExpiresIn:   cfg.AccessTokenExp,
	}, nil
}
