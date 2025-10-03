package ticket

import (
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