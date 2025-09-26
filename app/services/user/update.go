package user

import (
	"fmt"

	dbpkg "student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/openapi"
)

type UpdateFields struct {
	Email      *string
	Name       *string
	Phone      *string // "" => clear (NULL)
	Dept       *string // "" => clear (NULL)
	AllowEmail *bool
}

// ErrEmailTaken 当更新 email 与他人重复时
type ErrEmailTaken struct{ Email string }
func (e *ErrEmailTaken) Error() string { return fmt.Sprintf("邮箱已被占用: %s", e.Email) }

func (s *Service) UpdateByID(id uint, f UpdateFields) (*openapi.User, error) {
	// 取用户
	u, err := dbpkg.GetUserByID(s.db, id)
	if err != nil {
		return nil, err
	}

	// email 必填 & 唯一
	if f.Email == nil || *f.Email == "" {
		return nil, fmt.Errorf("email 为空")
	}
	if *f.Email != u.Email {
		taken, err := dbpkg.ExistsOtherUserWithEmail(s.db, *f.Email, u.ID)
		if err != nil {
			return nil, err
		}
		if taken {
			return nil, &ErrEmailTaken{Email: *f.Email}
		}
		u.Email = *f.Email
	}

	// 其他可选字段（仅当字段出现时才更新）
	if f.Name != nil {
		u.Name = *f.Name
	}
	if f.Phone != nil {
		if *f.Phone == "" {
			u.Phone = nil
		} else {
			u.Phone = f.Phone // 类型是 *string，直接赋
		}
	}
	if f.Dept != nil {
		if *f.Dept == "" {
			u.Dept = nil
		} else {
			u.Dept = f.Dept
		}
	}
	if f.AllowEmail != nil {
		u.AllowEmail = *f.AllowEmail
	}

	if err := dbpkg.UpdateUser(s.db, u); err != nil {
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