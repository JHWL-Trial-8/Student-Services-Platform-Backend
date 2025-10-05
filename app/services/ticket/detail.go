package ticket

import (
	"errors"
	dbpkg "student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/openapi"

	"gorm.io/gorm"
)

func (s *Service) GetTicketDetail(currentUID, ticketID uint) (*openapi.TicketDetail, error) {
	_, t, err := s.getTicketWithAccessCheck(currentUID, ticketID)
	if err != nil {
		return nil, err
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

	// 评分
	var rating *openapi.Rating
	var dbRating dbpkg.Rating
	if err := s.db.Where("ticket_id = ?", t.ID).First(&dbRating).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err // 实际查询错误
		}
		// 如果没有找到评分，rating 保持为 nil
	} else {
		rating = &openapi.Rating{
			Id:        int32(dbRating.ID),
			TicketId:  int32(dbRating.TicketID),
			UserId:    int32(dbRating.UserID),
			Stars:     int32(dbRating.Stars),
			Comment:   dbRating.Comment,
			CreatedAt: dbRating.CreatedAt,
		}
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
		// Messages 统一交给独立端点获取，此处不填充
		Messages: nil, 
		Rating:   rating, // 现在会正确加载或为 nil
	}

	return out, nil
}