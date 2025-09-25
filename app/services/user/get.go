package user

import (
	"student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/openapi"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

// NewService 连接一个 gorm DB
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetUserInfomationByID(id string) (*openapi.User, error) {
	var u db.User
	if err := s.db.Where("id = ?", id).First(&u).Error; err != nil {
		return nil, err
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
	return &apiUser, nil
}
