package ticket

import (
	dbpkg "student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/openapi"

	"errors"
	"gorm.io/gorm"
)

type ListFilters struct {
	Status       string // optional
	Category     string // optional
	IsUrgent     *bool  // optional
	AssignedToMe *bool  // admin only
}

// ListTickets 根据角色与筛选返回分页工单
func (s *Service) ListTickets(currentUID uint, f ListFilters, page, pageSize int) (*openapi.PagedTickets, error) {
	u, err := s.currentUser(s.db, currentUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &ErrForbidden{Reason: "user not found"}
		}
		return nil, err
	}

	q := s.db.Model(&dbpkg.Ticket{})
	// 权限：学生仅看自己的
	if !isAdmin(u.Role) {
		q = q.Where("user_id = ?", currentUID)
	} else {
		// 管理员：可筛选"我负责的"
		if f.AssignedToMe != nil && *f.AssignedToMe {
			q = q.Where("assigned_admin_id = ?", currentUID)
		}
	}

	if f.Status != "" {
		q = q.Where("status = ?", dbpkg.TicketStatus(f.Status))
	}
	if f.Category != "" {
		q = q.Where("category = ?", f.Category)
	}
	if f.IsUrgent != nil {
		q = q.Where("is_urgent = ?", *f.IsUrgent)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var rows []dbpkg.Ticket
	if err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]openapi.Ticket, 0, len(rows))
	for _, t := range rows {
		imgIDs, _ := dbpkg.GetTicketImageIDs(s.db, t.ID)
		img32 := make([]int32, 0, len(imgIDs))
		for _, id := range imgIDs {
			img32 = append(img32, int32(id))
		}
		items = append(items, openapi.Ticket{
			Id:              int32(t.ID),
			UserId:          int32(t.UserID),
			Title:           t.Title,
			Content:         t.Content,
			Category:        t.Category,
			IsUrgent:        t.IsUrgent,
			IsAnonymous:     t.IsAnonymous,
			Status:          openapi.TicketStatus(t.Status),
			AssignedAdminId: toPtrInt32FromUintPtr(t.AssignedAdminID),
			ClaimedAt:       t.ClaimedAt,
			CreatedAt:       t.CreatedAt,
			UpdatedAt:       t.UpdatedAt,
			ImageIds:        img32,
		})
	}

	return &openapi.PagedTickets{
		Items:    items,
		Page:     int32(page),
		PageSize: int32(pageSize),
		Total:    int32(total),
	}, nil
}
