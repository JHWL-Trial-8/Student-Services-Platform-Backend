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
	body := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>新工单通知</title>
	</head>
	<body>
		<h2>📋 新工单创建通知</h2>
		<p><strong>工单编号:</strong> #%d</p>
		<p><strong>工单标题:</strong> %s</p>
		<p><strong>工单分类:</strong> %s</p>
		<p><strong>创建人:</strong> %s</p>
		<p><strong>创建时间:</strong> %s</p>
		<hr>
		<p>请及时处理该工单。</p>
	</body>
	</html>
	`, ticketID, title, category, creatorName, time.Now().Format("2006-01-02 15:04:05"))

	emailContext := map[string]interface{}{
		"ticket_id": ticketID,
		"category":  category,
		"creator":   creatorName,
	}

	return n.emailService.SendEmailWithDynamicRecipients(ctx, worker.EmailTypeTicketCreated, subject, body, emailContext)
}

// NotifyTicketClaimed 通知工单被接单
func (n *Notifier) NotifyTicketClaimed(ctx context.Context, ticketID uint, title, handlerName, creatorEmail string) error {
	subject := fmt.Sprintf("工单已被接单 - %s", title)
	body := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>工单接单通知</title>
	</head>
	<body>
		<h2>✅ 您的工单已被接单</h2>
		<p><strong>工单编号:</strong> #%d</p>
		<p><strong>工单标题:</strong> %s</p>
		<p><strong>处理人:</strong> %s</p>
		<p><strong>接单时间:</strong> %s</p>
		<hr>
		<p>我们将尽快为您处理，请耐心等待。</p>
	</body>
	</html>
	`, ticketID, title, handlerName, time.Now().Format("2006-01-02 15:04:05"))

	emailContext := map[string]interface{}{
		"ticket_id":     ticketID,
		"handler_name":  handlerName,
		"creator_email": creatorEmail,
	}

	return n.emailService.SendEmailWithDynamicRecipients(ctx, worker.EmailTypeTicketClaimed, subject, body, emailContext)
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
	body := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>工单新消息通知</title>
	</head>
	<body>
		<h2>💬 工单收到新消息</h2>
		<p><strong>工单编号:</strong> #%d</p>
		<p><strong>发送人:</strong> %s</p>
		<p><strong>发送时间:</strong> %s</p>
		<div style="background: #e3f2fd; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #2196F3;">
			<h3>消息内容:</h3>
			<p>%s</p>
		</div>
		<hr>
		<p>请及时查看并回复。</p>
	</body>
	</html>
	`, ticketID, senderName, time.Now().Format("2006-01-02 15:04:05"), message)

	emailContext := map[string]interface{}{
		"ticket_id":     ticketID,
		"sender_name":   senderName,
		"message":       message,
		"creator_email": creatorEmail,
		"handler_email": handlerEmail,
	}

	return n.emailService.SendEmailWithDynamicRecipients(ctx, worker.EmailTypeMessageReceived, subject, body, emailContext)
}

// NotifyUserCreated 通知新用户创建
func (n *Notifier) NotifyUserCreated(ctx context.Context, userName, userEmail, userRole string) error {
	subject := fmt.Sprintf("新用户注册 - %s", userName)
	body := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>新用户注册通知</title>
	</head>
	<body>
		<h2>👤 新用户注册通知</h2>
		<p><strong>用户姓名:</strong> %s</p>
		<p><strong>用户邮箱:</strong> %s</p>
		<p><strong>用户角色:</strong> %s</p>
		<p><strong>注册时间:</strong> %s</p>
		<hr>
		<p>请关注新用户的使用情况。</p>
	</body>
	</html>
	`, userName, userEmail, userRole, time.Now().Format("2006-01-02 15:04:05"))

	emailContext := map[string]interface{}{
		"user_name":  userName,
		"user_email": userEmail,
		"user_role":  userRole,
	}

	return n.emailService.SendEmailWithDynamicRecipients(ctx, worker.EmailTypeUserCreated, subject, body, emailContext)
}

// NotifyPasswordReset 通知密码重置
func (n *Notifier) NotifyPasswordReset(ctx context.Context, userName, userEmail, resetToken string) error {
	subject := "密码重置请求"
	body := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>密码重置通知</title>
	</head>
	<body>
		<h2>🔒 密码重置请求</h2>
		<p>亲爱的 %s，</p>
		<p>我们收到了您的密码重置请求。</p>
		<p><strong>请求时间:</strong> %s</p>
		<div style="background: #fff3cd; padding: 15px; margin: 10px 0; border-radius: 5px; border: 1px solid #ffeaa7;">
			<p><strong>重置代码:</strong> <code style="background: #f1f1f1; padding: 5px;">%s</code></p>
			<p><em>此代码10分钟内有效。</em></p>
		</div>
		<hr>
		<p>如果这不是您的操作，请忽略此邮件。</p>
	</body>
	</html>
	`, userName, time.Now().Format("2006-01-02 15:04:05"), resetToken)

	emailContext := map[string]interface{}{
		"user_name":   userName,
		"user_email":  userEmail,
		"reset_token": resetToken,
	}

	return n.emailService.SendEmailWithDynamicRecipients(ctx, worker.EmailTypePasswordReset, subject, body, emailContext)
}

// NotifySystemMaintenance 通知系统维护
func (n *Notifier) NotifySystemMaintenance(ctx context.Context, title, description, startTime, endTime, level string) error {
	subject := fmt.Sprintf("系统维护通知 - %s", title)
	body := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>系统维护通知</title>
	</head>
	<body>
		<h2>🔧 系统维护通知</h2>
		<p><strong>维护标题:</strong> %s</p>
		<p><strong>维护级别:</strong> %s</p>
		<p><strong>开始时间:</strong> %s</p>
		<p><strong>结束时间:</strong> %s</p>
		<div style="background: #f8f9fa; padding: 15px; margin: 10px 0; border-radius: 5px;">
			<h3>维护说明:</h3>
			<p>%s</p>
		</div>
		<hr>
		<p>维护期间可能会影响系统使用，给您带来不便敬请谅解。</p>
	</body>
	</html>
	`, title, level, startTime, endTime, description)

	emailContext := map[string]interface{}{
		"title":             title,
		"description":       description,
		"start_time":        startTime,
		"end_time":          endTime,
		"maintenance_level": level,
	}

	return n.emailService.SendEmailWithDynamicRecipients(ctx, worker.EmailTypeSystemMaintenance, subject, body, emailContext)
}
