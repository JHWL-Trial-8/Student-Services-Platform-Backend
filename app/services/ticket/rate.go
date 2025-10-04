package ticket

import (
	"context"
	"fmt"
	"strings"
	"time"

	dbpkg "student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/openapi"

	errors "errors"
	"gorm.io/gorm"
)

// RateTicket 学生对自己的工单进行一次性评分
func (s *Service) RateTicket(currentUID, ticketID uint, stars int, comment string) (*openapi.Rating, error) {
	// 校验登录用户存在和工单权限
	_, t, err := s.getTicketWithAccessCheck(currentUID, ticketID)
	if err != nil {
		return nil, err
	}
	
	// 评分权限更严格：只能是创建者本人
	if t.UserID != currentUID {
		return nil, &ErrForbidden{Reason: "only ticket creator can rate"}
	}

	// 输入校验
	comment = strings.TrimSpace(comment)
	details := map[string]interface{}{}
	if stars < 1 || stars > 5 {
		details["stars"] = "须在 1-5 之间"
	}
	if l := len([]rune(comment)); l > 1000 {
		details["comment"] = "长度不能超过 1000"
	}
	if len(details) > 0 {
		return nil, &ErrValidation{Message: "字段校验失败", Details: details}
	}

	// 应用层"已评价"快速检查
	var existing dbpkg.Rating
	if err := s.db.Where("ticket_id = ?", ticketID).First(&existing).Error; err == nil {
		return nil, &ErrAlreadyRated{TicketID: ticketID}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 写入评分（依赖唯一索引 ticket_id 保证并发安全）
	now := time.Now().UTC().Truncate(time.Microsecond)
	r := &dbpkg.Rating{
		TicketID: ticketID,
		UserID:   currentUID,
		Stars:    stars,
		Comment:  comment,
		CreatedAt: now,
	}
	if err := s.db.Create(r).Error; err != nil {
		return nil, err
	}

	// 异步发送邮件通知给管理员（如果配置了 notifier）
	if s.notifier != nil {
		go func(ticketID uint, stars int, comment string) {
			// 获取工单和处理人的信息用于邮件
			var ticket dbpkg.Ticket
			var handler dbpkg.User

			if err := s.db.First(&ticket, ticketID).Error; err != nil {
				return // 静默失败，不影响主流程
			}
			
			// 检查是否有关联的管理员
			if ticket.AssignedAdminID != nil {
				if err := s.db.First(&handler, *ticket.AssignedAdminID).Error; err != nil {
					return
				}

				// 发送邮件通知
				s.notifier.NotifyTicketRated(
					context.Background(),
					ticketID,
					ticket.Title,
					fmt.Sprintf("%d星", stars),
					comment,
					handler.Email,
				)
			}
		}(ticketID, stars, comment)
	}

	return &openapi.Rating{
		Id:        int32(r.ID),
		TicketId:  int32(r.TicketID),
		UserId:    int32(r.UserID),
		Stars:     int32(r.Stars),
		Comment:   r.Comment,
		CreatedAt: r.CreatedAt,
	}, nil
}