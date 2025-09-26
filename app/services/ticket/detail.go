package ticket

import (
	"errors"

	dbpkg "student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/openapi"

	"gorm.io/gorm"
)

func (s *Service) GetTicketDetail(currentUID, ticketID uint) (*openapi.TicketDetail, error) {
	u, err := s.currentUser(s.db, currentUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &ErrForbidden{Reason: "user not found"}
		}
		return nil, err
	}

	// 取工单
	var t dbpkg.Ticket
	if err := s.db.First(&t, ticketID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &ErrNotFound{Resource: "ticket"}
		}
		return nil, err
	}

	// 权限：学生仅可看自己的
	if !isAdmin(u.Role) && t.UserID != currentUID {
		return nil, &ErrForbidden{Reason: "student cannot view others' ticket"}
	}

	// 图片
	imgIDs, err := dbpkg.GetTicketImageIDs(s.db, t.ID)
	if err != nil {
		return nil, err
	}
	img32 := make([]int32, 0, len(imgIDs))
	for _, id := range imgIDs {
		img32 = append(img32, int32(id))
	}

	out := &openapi.TicketDetail{
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
		// Messages 和 Rating 统一交给独立端点获取
	}

	return out, nil
}