package canned

import (
	"errors"
	"fmt"
	"strings"

	dbpkg "student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/openapi"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

func (s *Service) List(currentUID uint, page, pageSize int) (*openapi.PagedCannedReplies, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	q := s.db.Model(&dbpkg.CannedReply{}).
		Where("admin_user_id IN ?", []uint{0, currentUID})

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	var rows []dbpkg.CannedReply
	if err := q.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]openapi.CannedReply, 0, len(rows))
	for _, r := range rows {
		items = append(items, openapi.CannedReply{
			Id:          int32(r.ID),
			AdminUserId: int32(r.AdminUserID),
			Title:       r.Title,
			Body:        r.Body,
			CreatedAt:   r.CreatedAt,
		})
	}

	return &openapi.PagedCannedReplies{
		Items:    items,
		Page:     int32(page),
		PageSize: int32(pageSize),
		Total:    int32(total),
	}, nil
}

func validate(title, body string) error {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	if title == "" || body == "" {
		return fmt.Errorf("title 和 body 不能为空")
	}
	if len([]rune(title)) > 64 {
		return fmt.Errorf("title 过长（最多 64）")
	}
	if len([]rune(body)) > 2000 {
		return fmt.Errorf("body 过长（最多 2000）")
	}
	return nil
}

func (s *Service) Create(currentUID uint, in openapi.CannedReplyCreate) (*openapi.CannedReply, error) {
	if err := validate(in.Title, in.Body); err != nil {
		return nil, err
	}
	cr := &dbpkg.CannedReply{
		AdminUserID: currentUID,
		Title:       strings.TrimSpace(in.Title),
		Body:        strings.TrimSpace(in.Body),
	}
	if err := s.db.Create(cr).Error; err != nil {
		return nil, err
	}
	
	// 重新从数据库查询记录，确保获取到数据库中的时间戳
	var freshCR dbpkg.CannedReply
	if err := s.db.First(&freshCR, cr.ID).Error; err != nil {
		return nil, err
	}
	
	out := &openapi.CannedReply{
		Id:          int32(freshCR.ID),
		AdminUserId: int32(freshCR.AdminUserID),
		Title:       freshCR.Title,
		Body:        freshCR.Body,
		CreatedAt:   freshCR.CreatedAt,
	}
	return out, nil
}

func (s *Service) canModify(currentUID uint, cr *dbpkg.CannedReply) (bool, error) {
	// 所有者或超级管理员可以修改
	if cr.AdminUserID == currentUID {
		return true, nil
	}
	u, err := dbpkg.GetUserByID(s.db, currentUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, fmt.Errorf("无权限")
		}
		return false, err
	}
	if u.Role == dbpkg.RoleSuperAdmin {
		return true, nil
	}
	return false, fmt.Errorf("无权限")
}

func (s *Service) Update(currentUID, id uint, in openapi.CannedReplyUpdate) (*openapi.CannedReply, error) {
	var cr dbpkg.CannedReply
	if err := s.db.First(&cr, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("资源不存在")
		}
		return nil, err
	}
	if ok, err := s.canModify(currentUID, &cr); !ok {
		return nil, err
	}
	// 只更新提供的字段
	if in.Title != "" {
		if err := validate(in.Title, cr.Body); err != nil {
			return nil, err
		}
		cr.Title = strings.TrimSpace(in.Title)
	}
	if in.Body != "" {
		if err := validate(cr.Title, in.Body); err != nil {
			return nil, err
		}
		cr.Body = strings.TrimSpace(in.Body)
	}
	if err := s.db.Save(&cr).Error; err != nil {
		return nil, err
	}
	
	// 重新从数据库查询记录，确保获取到数据库中的时间戳
	var freshCR dbpkg.CannedReply
	if err := s.db.First(&freshCR, cr.ID).Error; err != nil {
		return nil, err
	}
	
	out := &openapi.CannedReply{
		Id:          int32(freshCR.ID),
		AdminUserId: int32(freshCR.AdminUserID),
		Title:       freshCR.Title,
		Body:        freshCR.Body,
		CreatedAt:   freshCR.CreatedAt,
	}
	return out, nil
}

func (s *Service) Delete(currentUID, id uint) error {
	var cr dbpkg.CannedReply
	if err := s.db.First(&cr, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("资源不存在")
		}
		return err
	}
	if ok, err := s.canModify(currentUID, &cr); !ok {
		return err
	}
	return s.db.Delete(&dbpkg.CannedReply{}, id).Error
}