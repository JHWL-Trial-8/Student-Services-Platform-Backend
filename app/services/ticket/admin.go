package ticket

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	dbpkg "student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/openapi"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// audit 记录审计日志
func (s *Service) audit(ctx context.Context, tx *gorm.DB, actorID uint, action, entity string, entityID uint, diff map[string]interface{}) error {
	var diffJSON datatypes.JSON
	if diff != nil {
		b, err := json.Marshal(diff)
		if err != nil {
			return fmt.Errorf("序列化diff失败: %w", err)
		}
		diffJSON = datatypes.JSON(b)
	}
	al := &dbpkg.AuditLog{
		ActorUserID: actorID,
		Action:      action,
		Entity:      entity,
		EntityID:    entityID,
		Diff:        diffJSON,
		CreatedAt:   time.Now().UTC(),
	}
	return tx.WithContext(ctx).Create(al).Error
}

// ClaimTicket 管理员接单（原子 CAS）
func (s *Service) ClaimTicket(ctx context.Context, adminUID, ticketID uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		result := tx.Model(&dbpkg.Ticket{}).
			Where("id = ? AND status = ?", ticketID, dbpkg.TicketStatusNew).
			Updates(map[string]interface{}{
				"assigned_admin_id": adminUID,
				"status":            dbpkg.TicketStatusClaimed,
				"claimed_at":        &now,
				"updated_at":        now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			var cur dbpkg.Ticket
			if err := tx.First(&cur, ticketID).Error; err != nil {
				return &ErrNotFound{Resource: "ticket"}
			}
			if cur.Status != dbpkg.TicketStatusNew {
				return &ErrInvalidState{Message: fmt.Sprintf("仅 'NEW' 状态的工单可被认领, 当前为 '%s'", cur.Status)}
			}
			return &ErrConflict{Message: "工单已被他人认领"}
		}
		diff := map[string]interface{}{"status_to": "CLAIMED", "assigned_admin_id": adminUID}
		return s.audit(ctx, tx, adminUID, "ticket.claim", "TICKET", ticketID, diff)
	})
}

// UnclaimTicket 管理员撤销接单（原子 CAS）
func (s *Service) UnclaimTicket(ctx context.Context, adminUID, ticketID uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&dbpkg.Ticket{}).
			Where("id = ? AND assigned_admin_id = ? AND status = ?", ticketID, adminUID, dbpkg.TicketStatusClaimed).
			Updates(map[string]interface{}{
				"assigned_admin_id": gorm.Expr("NULL"),
				"status":            dbpkg.TicketStatusNew,
				"claimed_at":        gorm.Expr("NULL"),
				"updated_at":        time.Now().UTC(),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			var cur dbpkg.Ticket
			if err := tx.First(&cur, ticketID).Error; err != nil {
				return &ErrNotFound{Resource: "ticket"}
			}
			if cur.AssignedAdminID == nil || *cur.AssignedAdminID != adminUID {
				return &ErrForbidden{Reason: "你不是该工单的负责人"}
			}
			return &ErrInvalidState{Message: fmt.Sprintf("仅 'CLAIMED' 状态可撤销, 当前为 '%s'", cur.Status)}
		}
		diff := map[string]interface{}{"status_to": "NEW", "unassigned_admin_id": adminUID}
		return s.audit(ctx, tx, adminUID, "ticket.unclaim", "TICKET", ticketID, diff)
	})
}

// updateTicketStatusAsAdmin 通用的管理员状态更新函数
func (s *Service) updateTicketStatusAsAdmin(ctx context.Context, adminUID, ticketID uint, newStatus dbpkg.TicketStatus, action string, allowedOldStatuses []dbpkg.TicketStatus) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&dbpkg.Ticket{}).
			Where("id = ? AND assigned_admin_id = ? AND status IN ?", ticketID, adminUID, allowedOldStatuses).
			Updates(map[string]interface{}{
				"status":     newStatus,
				"updated_at": time.Now().UTC(),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			var cur dbpkg.Ticket
			if err := tx.First(&cur, ticketID).Error; err != nil {
				return &ErrNotFound{Resource: "ticket"}
			}
			if cur.AssignedAdminID == nil || *cur.AssignedAdminID != adminUID {
				return &ErrForbidden{Reason: "你不是该工单的负责人"}
			}
			return &ErrInvalidState{Message: fmt.Sprintf("工单当前状态 ('%s') 无法执行此操作", cur.Status)}
		}
		diff := map[string]interface{}{"status_to": string(newStatus)}
		return s.audit(ctx, tx, adminUID, action, "TICKET", ticketID, diff)
	})
}

// ResolveTicket 标记工单为已处理
func (s *Service) ResolveTicket(ctx context.Context, adminUID, ticketID uint) error {
	allowed := []dbpkg.TicketStatus{dbpkg.TicketStatusClaimed, dbpkg.TicketStatusInProgress}
	return s.updateTicketStatusAsAdmin(ctx, adminUID, ticketID, dbpkg.TicketStatusResolved, "ticket.resolve", allowed)
}

// CloseTicket 关闭工单 (负责人或超管)
func (s *Service) CloseTicket(ctx context.Context, adminUID, ticketID uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var u dbpkg.User
		if err := tx.Select("role").First(&u, adminUID).Error; err != nil {
			return &ErrForbidden{Reason: "user not found"}
		}
		var t dbpkg.Ticket
		if err := tx.First(&t, ticketID).Error; err != nil {
			return &ErrNotFound{Resource: "ticket"}
		}
		if !isSuperAdmin(u.Role) && (t.AssignedAdminID == nil || *t.AssignedAdminID != adminUID) {
			return &ErrForbidden{Reason: "只有负责人或超级管理员可以关闭工单"}
		}
		if t.Status != dbpkg.TicketStatusResolved {
			return &ErrInvalidState{Message: fmt.Sprintf("仅 'RESOLVED' 状态的工单可关闭, 当前为 '%s'", t.Status)}
		}
		if err := tx.Model(&t).Update("status", dbpkg.TicketStatusClosed).Error; err != nil {
			return err
		}
		diff := map[string]interface{}{"status_to": "CLOSED"}
		return s.audit(ctx, tx, adminUID, "ticket.close", "TICKET", ticketID, diff)
	})
}

// SpamFlag 管理员标记垃圾（进入待审）
func (s *Service) SpamFlag(ctx context.Context, adminUID, ticketID uint, reason string) (*openapi.SpamFlag, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, &ErrValidation{Message: "字段校验失败", Details: map[string]interface{}{"reason": "必填"}}
	}

	var sf dbpkg.SpamFlag
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var t dbpkg.Ticket
		if err := tx.First(&t, ticketID).Error; err != nil {
			return &ErrNotFound{Resource: "ticket"}
		}
		if t.Status == dbpkg.TicketStatusSpamPending || t.Status == dbpkg.TicketStatusSpamConfirmed {
			return &ErrConflict{Message: "该工单已被标记为垃圾"}
		}
		sf = dbpkg.SpamFlag{TicketID: ticketID}
		sf.FlaggedByAdminID = adminUID
		sf.Reason = reason
		sf.Status = "PENDING"
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "ticket_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"flagged_by_admin_id", "reason", "status", "reviewed_by_super_admin_id", "reviewed_at"}),
		}).Create(&sf).Error; err != nil {
			return err
		}
		oldStatus := t.Status
		if err := tx.Model(&t).Update("status", dbpkg.TicketStatusSpamPending).Error; err != nil {
			return err
		}
		diff := map[string]interface{}{"status_from": oldStatus, "status_to": "SPAM_PENDING", "reason": reason}
		return s.audit(ctx, tx, adminUID, "ticket.spam_flag", "TICKET", ticketID, diff)
	})
	if err != nil {
		return nil, err
	}
	return &openapi.SpamFlag{
		TicketId: int32(sf.TicketID), FlaggedByAdminId: int32(sf.FlaggedByAdminID), Reason: sf.Reason, Status: sf.Status}, nil
}

// SpamReview 超管审核垃圾（approve/reject）
func (s *Service) SpamReview(ctx context.Context, superAdminUID, ticketID uint, action string) error {
	act := strings.ToLower(strings.TrimSpace(action))
	if act != "approve" && act != "reject" {
		return &ErrValidation{Message: "字段校验失败", Details: map[string]interface{}{"action": "必须为 'approve' 或 'reject'"}}
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ticketTo dbpkg.TicketStatus
		var spamStatus string
		if act == "approve" {
			ticketTo = dbpkg.TicketStatusSpamConfirmed
			spamStatus = "CONFIRMED"
		} else {
			ticketTo = dbpkg.TicketStatusSpamRejected
			spamStatus = "REJECTED"
		}
		
		// 获取工单信息以获取学生ID
		var ticket dbpkg.Ticket
		if err := tx.First(&ticket, ticketID).Error; err != nil {
			return &ErrNotFound{Resource: "ticket"}
		}
		
		res := tx.Model(&dbpkg.Ticket{}).Where("id = ? AND status = ?", ticketID, dbpkg.TicketStatusSpamPending).Update("status", ticketTo)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return &ErrInvalidState{Message: "仅 'SPAM_PENDING' 状态的工单可审核"}
		}
		now := time.Now().UTC()
		if err := tx.Model(&dbpkg.SpamFlag{}).Where("ticket_id = ?", ticketID).Updates(map[string]interface{}{
			"status": spamStatus, "reviewed_by_super_admin_id": superAdminUID, "reviewed_at": &now}).Error; err != nil {
			return err
		}
		
		// 如果审核通过，发送自动回复消息
		if act == "approve" {
			autoMsg := "请您在提交反馈时确保内容的有效性和准确性，感谢您的理解和配合。如有异议，请重新反馈。"
			msg := &dbpkg.TicketMessage{
				TicketID:       ticket.ID,
				SenderUserID:   superAdminUID, // 使用超级管理员ID作为发送者
				Body:           autoMsg,
				IsInternalNote: false, // 学生可见
				CreatedAt:      now,
			}
			if err := tx.Create(msg).Error; err != nil {
				return err
			}
		}
		
		diff := map[string]interface{}{"status_to": string(ticketTo), "review_action": act}
		return s.audit(ctx, tx, superAdminUID, "ticket.spam_review", "TICKET", ticketID, diff)
	})
}

// isSuperAdmin 检查用户是否为超级管理员
func isSuperAdmin(role dbpkg.Role) bool {
	return role == dbpkg.RoleSuperAdmin
}