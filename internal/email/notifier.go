package email

import (
	"context"
	"fmt"
	"time"

	"student-services-platform-backend/internal/worker"
)

// Notifier 邮件通知器，提供便捷的业务邮件发送方法
type Notifier struct {
	emailService *Service
}

// NewNotifier 创建邮件通知器
func NewNotifier(emailService *Service) *Notifier {
	return &Notifier{
		emailService: emailService,
	}
}

// NotifyTicketCreated 通知工单创建
func (n *Notifier) NotifyTicketCreated(ctx context.Context, ticketID uint, title, category, creatorName string) error {
	subject := fmt.Sprintf("新工单创建 - %s", title)

	emailContext := map[string]interface{}{
		"ticket_id":    ticketID,
		"title":        title,
		"category":     category,
		"student_name": creatorName,
		"created_at":   time.Now().Format("2006-01-02 15:04:05"),
	}

	return n.emailService.SendEmailWithDynamicRecipients(ctx, worker.EmailTypeTicketCreated, subject, "", emailContext)
}

// NotifyTicketClaimed 通知工单被接单
func (n *Notifier) NotifyTicketClaimed(ctx context.Context, ticketID uint, title, handlerName, creatorEmail string) error {
	subject := fmt.Sprintf("工单已被接单 - %s", title)

	emailContext := map[string]interface{}{
		"ticket_id":     ticketID,
		"title":         title,
		"admin_name":    handlerName,
		"student_name":  "学生", // 默认值，实际应用中应从数据库获取
		"claimed_at":    time.Now().Format("2006-01-02 15:04:05"),
		"student_email": creatorEmail,
		"ticket_url":    fmt.Sprintf("/tickets/%d", ticketID), // 前端路由
	}

	return n.emailService.SendEmailWithDynamicRecipients(ctx, worker.EmailTypeTicketClaimed, subject, "", emailContext)
}

// NotifyTicketResolved 通知工单已处理
func (n *Notifier) NotifyTicketResolved(ctx context.Context, ticketID uint, title, resolution, handlerName, creatorEmail, handlerEmail string) error {
	// 使用模板发送邮件，而不是硬编码HTML
	emailContext := map[string]interface{}{
		"ticket_id":     ticketID,
		"title":         title,
		"resolution":    resolution,
		"admin_name":    handlerName,
		"student_name":  "学生", // 默认值，实际应用中应从数据库获取
		"resolved_at":   time.Now().Format("2006-01-02 15:04:05"),
		"creator_email": creatorEmail,
		"handler_email": handlerEmail,
		"ticket_url":    fmt.Sprintf("/tickets/%d", ticketID), // 前端路由
	}

	// 使用动态收件人解析器
	return n.emailService.SendEmailWithDynamicRecipients(
		ctx,
		worker.EmailTypeTicketResolved,
		fmt.Sprintf("工单已处理 - %s", title),
		"", // 空的body，使用模板
		emailContext,
	)
}

// NotifyTicketClosed 通知工单已关闭
func (n *Notifier) NotifyTicketClosed(ctx context.Context, ticketID uint, title, handlerName, creatorEmail, handlerEmail string) error {
	// 使用模板发送邮件
	emailContext := map[string]interface{}{
		"ticket_id":     ticketID,
		"title":         title,
		"admin_name":    handlerName,
		"student_name":  "学生", // 默认值，实际应用中应从数据库获取
		"closed_at":     time.Now().Format("2006-01-02 15:04:05"),
		"creator_email": creatorEmail,
		"handler_email": handlerEmail,
		"ticket_url":    fmt.Sprintf("/tickets/%d", ticketID), // 前端路由
	}

	// 使用动态收件人解析器
	return n.emailService.SendEmailWithDynamicRecipients(
		ctx,
		worker.EmailTypeTicketClosed,
		fmt.Sprintf("工单已关闭 - %s", title),
		"", // 空的body，使用模板
		emailContext,
	)
}

// NotifyNewMessage 通知收到新消息
func (n *Notifier) NotifyNewMessage(ctx context.Context, ticketID uint, senderName, message, creatorEmail, handlerEmail string) error {
	subject := fmt.Sprintf("工单新消息 - #%d", ticketID)

	emailContext := map[string]interface{}{
		"ticket_id":      ticketID,
		"title":          "工单标题", // 默认值，实际应用中应从数据库获取
		"sender_name":    senderName,
		"message_body":   message,
		"recipient_name": "用户", // 默认值，实际应用中应确定收件人姓名
		"message_time":   time.Now().Format("2006-01-02 15:04:05"),
		"creator_email":  creatorEmail,
		"handler_email":  handlerEmail,
		"ticket_url":     fmt.Sprintf("/tickets/%d", ticketID), // 前端路由
	}

	return n.emailService.SendEmailWithDynamicRecipients(ctx, worker.EmailTypeMessageReceived, subject, "", emailContext)
}

// NotifyUserCreated 通知新用户创建
func (n *Notifier) NotifyUserCreated(ctx context.Context, userName, userEmail, userRole string) error {
	subject := fmt.Sprintf("新用户注册 - %s", userName)

	emailContext := map[string]interface{}{
		"user_name":  userName,
		"user_email": userEmail,
		"user_role":  userRole,
		"created_at": time.Now().Format("2006-01-02 15:04:05"),
	}

	return n.emailService.SendEmailWithDynamicRecipients(ctx, worker.EmailTypeUserCreated, subject, "", emailContext)
}

// NotifyPasswordReset 通知密码重置
func (n *Notifier) NotifyPasswordReset(ctx context.Context, userName, userEmail, resetToken string) error {
	subject := "密码重置请求"

	emailContext := map[string]interface{}{
		"user_name":    userName,
		"user_email":   userEmail,
		"reset_token":  resetToken,
		"request_time": time.Now().Format("2006-01-02 15:04:05"),
	}

	return n.emailService.SendEmailWithDynamicRecipients(ctx, worker.EmailTypePasswordReset, subject, "", emailContext)
}

// NotifySystemMaintenance 通知系统维护
func (n *Notifier) NotifySystemMaintenance(ctx context.Context, title, description, startTime, endTime, level string) error {
	subject := fmt.Sprintf("系统维护通知 - %s", title)

	emailContext := map[string]interface{}{
		"title":             title,
		"description":       description,
		"start_time":        startTime,
		"end_time":          endTime,
		"maintenance_level": level,
		"notice_time":       time.Now().Format("2006-01-02 15:04:05"),
	}

	return n.emailService.SendEmailWithDynamicRecipients(ctx, worker.EmailTypeSystemMaintenance, subject, "", emailContext)
}

// NotifyTicketUnclaimed 通知工单被撤销
func (n *Notifier) NotifyTicketUnclaimed(ctx context.Context, ticketID uint, title, creatorEmail string) error {
	subject := fmt.Sprintf("工单已被撤销 - %s", title)

	emailContext := map[string]interface{}{
		"ticket_id":     ticketID,
		"title":         title,
		"student_email": creatorEmail,
		"unclaimed_at":  time.Now().Format("2006-01-02 15:04:05"),
		"ticket_url":    fmt.Sprintf("/tickets/%d", ticketID),
	}

	return n.emailService.SendEmailWithDynamicRecipients(ctx, worker.EmailTypeTicketUnclaimed, subject, "", emailContext)
}

// NotifyTicketRated 通知工单被评价
func (n *Notifier) NotifyTicketRated(ctx context.Context, ticketID uint, title, rating string, comments, handlerEmail string) error {
	subject := fmt.Sprintf("工单已被评价 - %s", title)

	emailContext := map[string]interface{}{
		"ticket_id":     ticketID,
		"title":         title,
		"rating":        rating,
		"comments":      comments,
		"handler_email": handlerEmail,
		"rated_at":      time.Now().Format("2006-01-02 15:04:05"),
		"ticket_url":    fmt.Sprintf("/tickets/%d", ticketID),
	}

	return n.emailService.SendEmailWithDynamicRecipients(ctx, worker.EmailTypeTicketRated, subject, "", emailContext)
}

// NotifySpamFlagged 通知垃圾标记
func (n *Notifier) NotifySpamFlagged(ctx context.Context, ticketID uint, title, reporterName string) error {
	subject := fmt.Sprintf("工单被标记为垃圾 - %s", title)

	emailContext := map[string]interface{}{
		"ticket_id":    ticketID,
		"title":        title,
		"reporter_name": reporterName,
		"flagged_at":   time.Now().Format("2006-01-02 15:04:05"),
		"ticket_url":   fmt.Sprintf("/tickets/%d", ticketID),
	}

	return n.emailService.SendEmailWithDynamicRecipients(ctx, worker.EmailTypeSpamFlagged, subject, "", emailContext)
}

// NotifySpamReviewed 通知垃圾审核结果
func (n *Notifier) NotifySpamReviewed(ctx context.Context, ticketID uint, title, creatorEmail, result string) error {
	subject := fmt.Sprintf("垃圾审核结果 - %s", title)

	emailContext := map[string]interface{}{
		"ticket_id":     ticketID,
		"title":         title,
		"creator_email": creatorEmail,
		"result":        result, // "已确认" 或 "误报"
		"reviewed_at":   time.Now().Format("2006-01-02 15:04:05"),
		"ticket_url":    fmt.Sprintf("/tickets/%d", ticketID),
	}

	return n.emailService.SendEmailWithDynamicRecipients(ctx, worker.EmailTypeSpamReviewed, subject, "", emailContext)
}
