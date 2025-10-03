package worker

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// EmailService 邮件服务接口
type EmailService interface {
	SendEmail(ctx context.Context, task *EmailTask) error
	SendTemplateEmail(ctx context.Context, emailType string, context map[string]interface{}) error
}

// EmailHandler 邮件任务处理器
type EmailHandler struct {
	emailService EmailService
}

// NewEmailHandler 创建邮件处理器
func NewEmailHandler(emailService EmailService) *EmailHandler {
	return &EmailHandler{
		emailService: emailService,
	}
}

// HandleEmailTask 处理邮件发送任务
func (h *EmailHandler) HandleEmailTask(ctx context.Context, task *EmailTask) error {
	log.Printf("处理邮件任务: type=%s, to=%v, subject=%s",
		task.Type, task.To, task.Subject)

	// 根据邮件类型选择处理方式
	var err error
	if h.needsTemplate(task.Type) {
		err = h.emailService.SendTemplateEmail(ctx, string(task.Type), task.Context)
	} else {
		err = h.emailService.SendEmail(ctx, task)
	}

	if err != nil {
		log.Printf("邮件发送失败: %v", err)
		// 对于某些类型的错误，不进行重试
		if isNonRetryableError(err) {
			log.Printf("不可重试的错误，放弃任务: %v", err)
			return fmt.Errorf("不可重试的错误: %w", err)
		}
		return fmt.Errorf("邮件发送失败: %w", err)
	}

	log.Printf("邮件发送成功: type=%s, to=%v", task.Type, task.To)
	return nil
}

// needsTemplate 判断是否需要使用模板
func (h *EmailHandler) needsTemplate(emailType EmailType) bool {
	templateTypes := map[EmailType]bool{
		EmailTypeTicketCreated:     true,
		EmailTypeTicketClaimed:     true,
		EmailTypeTicketUnclaimed:   true,
		EmailTypeTicketResolved:    true,
		EmailTypeTicketClosed:      true,
		EmailTypeTicketRated:       true,
		EmailTypeMessageReceived:   true,
		EmailTypeSpamFlagged:       true,
		EmailTypeSpamReviewed:      true,
		EmailTypeUserCreated:       true,
		EmailTypePasswordReset:     true,
		EmailTypeSystemMaintenance: true,
	}

	return templateTypes[emailType]
}

// isNonRetryableError 判断是否为不可重试的错误
func isNonRetryableError(err error) bool {
	errorMsg := err.Error()
	// 认证相关错误不重试
	if strings.Contains(errorMsg, "535 authentication failed") {
		return true
	}
	// 邮箱不存在
	if strings.Contains(errorMsg, "550 mailbox unavailable") {
		return true
	}
	// 其他不可恢复的错误可以在这里添加
	return false
}
