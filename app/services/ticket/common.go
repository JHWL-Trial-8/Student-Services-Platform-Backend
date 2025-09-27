package ticket

import (
	"errors"
	"fmt"
	dbpkg "student-services-platform-backend/internal/db"
	"gorm.io/gorm"
)

// Service 封装工单领域逻辑
type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// ---- 共享错误类型 ----

type ErrValidation struct {
	Message string
	Details map[string]interface{}
}

func (e *ErrValidation) Error() string { return e.Message }

type ErrImageNotFound struct{ Missing []uint }

func (e *ErrImageNotFound) Error() string { return fmt.Sprintf("图片不存在: %v", e.Missing) }

type ErrForbidden struct{ Reason string }

func (e *ErrForbidden) Error() string { return "forbidden: " + e.Reason }

type ErrNotFound struct{ Resource string }

func (e *ErrNotFound) Error() string { return "not found: " + e.Resource }

type ErrAlreadyRated struct{ TicketID uint }

func (e *ErrAlreadyRated) Error() string { return "already rated" }

type ErrInvalidState struct{ Message string }

func (e *ErrInvalidState) Error() string { return "invalid state: " + e.Message }

type ErrConflict struct{ Message string }

func (e *ErrConflict) Error() string { return "conflict: " + e.Message }

// ---- 共享辅助函数 ----

func (s *Service) currentUser(db *gorm.DB, uid uint) (*dbpkg.User, error) {
	return dbpkg.GetUserByID(db, uid)
}

func isAdmin(role dbpkg.Role) bool {
	return role == dbpkg.RoleAdmin || role == dbpkg.RoleSuperAdmin
}

func toPtrInt32FromUintPtr(p *uint) *int32 {
	if p == nil {
		return nil
	}
	v := int32(*p)
	return &v
}

// getTicketWithAccessCheck 是一个核心的内部辅助函数。
// 它获取用户和工单，并检查当前用户是否有权访问该工单。
// 权限规则：学生只能访问自己的工单，管理员/超管可以访问任何工单。
func (s *Service) getTicketWithAccessCheck(currentUID, ticketID uint) (*dbpkg.User, *dbpkg.Ticket, error) {
	u, err := s.currentUser(s.db, currentUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, &ErrForbidden{Reason: "user not found"}
		}
		return nil, nil, err
	}

	var t dbpkg.Ticket
	if err := s.db.First(&t, ticketID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, &ErrNotFound{Resource: "ticket"}
		}
		return nil, nil, err
	}

	// 核心权限检查
	if !isAdmin(u.Role) && t.UserID != currentUID {
		return nil, nil, &ErrForbidden{Reason: "student cannot access others' ticket"}
	}

	return u, &t, nil
}