package user

import (
	"student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/openapi"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func (s *Service) GetUserInfomationByID(id string) (*openapi.User, error) {
	user := db.User{}
	result := s.db.Where("ID = ?", id).First(&user)
	if result.Error != nil {
		apiUser := openapi.User{
			Id:         int32(user.ID),
			Email:      user.Email,
			Name:       user.Email,
			Role:       openapi.Role(user.Role),
			Phone:      user.Phone,
			Dept:       user.Dept,
			IsActive:   user.IsActive,
			AllowEmail: user.AllowEmail,
			CreatedAt:  user.CreatedAt,
			UpdatedAt:  user.UpdatedAt,
		}
		return &apiUser, nil
	}
	return nil, result.Error

}
