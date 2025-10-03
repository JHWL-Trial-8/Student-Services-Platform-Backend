package ticket

import (
	"context"
	"strings"
	"time"

	dbpkg "student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/openapi"
)

func (s *Service) ListMessages(currentUID, ticketID uint, page, pageSize int) (*openapi.PagedTicketMessages, error) {
	u, _, err := s.getTicketWithAccessCheck(currentUID, ticketID)
	if err != nil {
		return nil, err
	}

	q := s.db.Model(&dbpkg.TicketMessage{}).Where("ticket_id = ?", ticketID)
	// 学生：看不到内部备注
	if !isAdmin(u.Role) {
		q = q.Where("is_internal_note = ?", false)
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

	var rows []dbpkg.TicketMessage
	if err := q.Order("id ASC").Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]openapi.TicketMessage, 0, len(rows))
	for _, m := range rows {
		items = append(items, openapi.TicketMessage{
			Id:             int32(m.ID),
			TicketId:       int32(m.TicketID),
			SenderUserId:   int32(m.SenderUserID),
			Body:           m.Body,
			IsInternalNote: m.IsInternalNote,
			CreatedAt:      m.CreatedAt,
		})
	}

	return &openapi.PagedTicketMessages{
		Items:    items,
		Page:     int32(page),
		PageSize: int32(pageSize),
		Total:    int32(total),
	}, nil
}

func (s *Service) PostMessage(currentUID, ticketID uint, body string, isInternal bool) (*openapi.TicketMessage, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, &ErrValidation{
			Message: "字段校验失败",
			Details: map[string]interface{}{"body": "必填"},
		}
	}

	u, t, err := s.getTicketWithAccessCheck(currentUID, ticketID)
	if err != nil {
		return nil, err
	}

	// 权限：学生不能发内部备注
	if !isAdmin(u.Role) && isInternal {
		return nil, &ErrForbidden{Reason: "student cannot post internal note"}
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	m := &dbpkg.TicketMessage{
		TicketID:       t.ID,
		SenderUserID:   currentUID,
		Body:           body,
		IsInternalNote: isInternal && isAdmin(u.Role),
		CreatedAt:      now,
	}
	if err := s.db.Create(m).Error; err != nil {
		return nil, err
	}

	// 发送邮件通知（如果配置了notifier）
	if s.notifier != nil {
		go func() {
			// 获取发送者和接收者信息用于邮件
			var sender dbpkg.User
			var creator dbpkg.User
			var handler dbpkg.User

			if err := s.db.First(&sender, currentUID).Error; err != nil {
				return // 静默失败，不影响主流程
			}

			if err := s.db.First(&creator, t.UserID).Error; err != nil {
				return
			}

			// 如果工单有处理人，获取处理人信息
			if t.AssignedAdminID != nil {
				if err := s.db.First(&handler, *t.AssignedAdminID).Error; err != nil {
					return
				}
			}

			// 确定收件人
			var recipientEmail string
			
			// 如果当前用户是学生，通知处理人
			if !isAdmin(u.Role) && t.AssignedAdminID != nil {
				recipientEmail = handler.Email
			} else if isAdmin(u.Role) {
				// 如果当前用户是管理员，通知学生
				recipientEmail = creator.Email
			}

			if recipientEmail != "" {
				// 发送新消息通知
				s.notifier.NotifyNewMessage(
					context.Background(),
					t.ID,
					sender.Name,
					body,
					creator.Email,
					handler.Email,
				)
			}
		}()
	}

	out := &openapi.TicketMessage{
		Id:             int32(m.ID),
		TicketId:       int32(m.TicketID),
		SenderUserId:   int32(m.SenderUserID),
		Body:           m.Body,
		IsInternalNote: m.IsInternalNote,
		CreatedAt:      m.CreatedAt,
	}
	return out, nil
}