package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	dbpkg "student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/openapi"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	db  *gorm.DB
	cfg *JWTConfig
}

type JWTConfig struct {
	SecretKey      string
	AccessTokenExp time.Duration
	Issuer         string
	Audience       string
}

func NewService(db *gorm.DB, cfg *JWTConfig) *Service {
	return &Service{db: db, cfg: cfg}
}

// Errors
type ErrUserNotFound struct{ Email string }
func (e *ErrUserNotFound) Error() string { return fmt.Sprintf("用户不存在: %s", e.Email) }

type ErrInvalidPassword struct{ Email string }
func (e *ErrInvalidPassword) Error() string { return fmt.Sprintf("密码错误: %s", e.Email) }

type ErrGenerateToken struct{ Message string }
func (e *ErrGenerateToken) Error() string { return fmt.Sprintf("生成令牌失败: %s", e.Message) }

type ErrEmailTaken struct{ Email string }
func (e *ErrEmailTaken) Error() string { return fmt.Sprintf("邮箱已被占用: %s", e.Email) }

// Register creates a user with a bcrypt hash.
func (s *Service) Register(req openapi.UserCreate) (*openapi.User, error) {
	// Uniqueness check (DB also enforces unique index)
	if _, err := dbpkg.GetUserByEmail(s.db, req.Email); err == nil {
		return nil, &ErrEmailTaken{Email: req.Email}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("生成密码哈希失败: %w", err)
	}

	u := &dbpkg.User{
		Email:        req.Email,
		Name:         req.Name,
		Role:         dbpkg.Role(req.Role), // "STUDENT"/"ADMIN"/"SUPER_ADMIN"
		Phone:        nilIfEmpty(req.Phone),
		Dept:         nilIfEmpty(req.Dept),
		IsActive:     true,
		AllowEmail:   true,
		PasswordHash: string(hash),
	}

	if err := dbpkg.CreateUser(s.db, u); err != nil {
		// likely unique conflict race
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return &openapi.User{
		Id:         int32(u.ID),
		Email:      u.Email,
		Name:       u.Name,
		Role:       openapi.Role(u.Role),
		Phone:      u.Phone,
		Dept:       u.Dept,
		IsActive:   u.IsActive,
		AllowEmail: u.AllowEmail,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}, nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (s *Service) Login(email, password string) (*openapi.AuthLoginPost200Response, error) {
	u, err := dbpkg.GetUserByEmail(s.db, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &ErrUserNotFound{Email: email}
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, &ErrInvalidPassword{Email: email}
	}

	tokenResp, err := s.generateAccessToken(u)
	if err != nil {
		return nil, &ErrGenerateToken{Message: err.Error()}
	}

	return &openapi.AuthLoginPost200Response{
		AccessToken: tokenResp.AccessToken,
		TokenType:   "bearer",
		ExpiresIn:   int32(tokenResp.ExpiresIn.Seconds()),
	}, nil
}

func (s *Service) generateAccessToken(u *dbpkg.User) (*struct {
	AccessToken string        `json:"access_token"`
	ExpiresIn   time.Duration `json:"expires_in"`
}, error) {
	if s.cfg.SecretKey == "" {
		return nil, errors.New("JWT 密钥未配置")
	}
	if s.cfg.AccessTokenExp <= 0 {
		return nil, errors.New("JWT 访问令牌有效期无效")
	}

	apiUser := openapi.User{
		Id:         int32(u.ID),
		Email:      u.Email,
		Name:       u.Name,
		Role:       openapi.Role(u.Role),
		Phone:      u.Phone,
		Dept:       u.Dept,
		IsActive:   u.IsActive,
		AllowEmail: u.AllowEmail,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}

	subject := strconv.Itoa(int(apiUser.Id))
	claims := jwt.RegisteredClaims{
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.AccessTokenExp)),
		Issuer:    s.cfg.Issuer,
		Audience:  []string{s.cfg.Audience},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.SecretKey))
	if err != nil {
		return nil, fmt.Errorf("JWT 签名失败: %w", err)
	}

	return &struct {
		AccessToken string        `json:"access_token"`
		ExpiresIn   time.Duration `json:"expires_in"`
	}{
		AccessToken: tokenString,
		ExpiresIn:   s.cfg.AccessTokenExp,
	}, nil
}