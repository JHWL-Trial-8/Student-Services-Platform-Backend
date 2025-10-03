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
		"claimed_at":    time.Now().Format("2006-01-02 15:04:05"),
		"creator_email": creatorEmail,
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
		"student_name":  "", // 这里需要从数据库查询学生姓名
		"resolved_at":   time.Now().Format("2006-01-02 15:04:05"),
		"creator_email": creatorEmail,
		"handler_email": handlerEmail,
		"ticket_url":    fmt.Sprintf("#/tickets/%d", ticketID), // 前端路由
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

// NotifyNewMessage 通知收到新消息
func (n *Notifier) NotifyNewMessage(ctx context.Context, ticketID uint, senderName, message, creatorEmail, handlerEmail string) error {
	subject := fmt.Sprintf("工单新消息 - #%d", ticketID)

	emailContext := map[string]interface{}{
		"ticket_id":     ticketID,
		"sender_name":   senderName,
		"message_body":  message,
		"message_time":  time.Now().Format("2006-01-02 15:04:05"),
		"creator_email": creatorEmail,
		"handler_email": handlerEmail,
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
	}

	return n.emailService.SendEmailWithDynamicRecipients(ctx, worker.EmailTypeSystemMaintenance, subject, "", emailContext)
}
