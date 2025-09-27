package email

import (
	"context"
	"fmt"

	"student-services-platform-backend/internal/worker"
)

// RecipientResolver 收件人解析器接口
type RecipientResolver interface {
	// ResolveRecipients 根据邮件类型和上下文解析收件人
	ResolveRecipients(ctx context.Context, emailType worker.EmailType, context map[string]interface{}) ([]string, error)
}

// DefaultRecipientResolver 默认收件人解析器
type DefaultRecipientResolver struct {
	// 使用发件人邮箱作为默认管理员邮箱
	defaultAdminEmail string
}

// NewDefaultRecipientResolver 创建默认收件人解析器
func NewDefaultRecipientResolver(defaultAdminEmail string) *DefaultRecipientResolver {
	return &DefaultRecipientResolver{
		defaultAdminEmail: defaultAdminEmail,
	}
}

// ResolveRecipients 实现收件人解析逻辑
func (r *DefaultRecipientResolver) ResolveRecipients(ctx context.Context, emailType worker.EmailType, emailContext map[string]interface{}) ([]string, error) {
	switch emailType {
	case worker.EmailTypeTicketCreated:
		return r.resolveTicketCreatedRecipients(ctx, emailContext)
	case worker.EmailTypeTicketClaimed:
		return r.resolveTicketClaimedRecipients(ctx, emailContext)
	case worker.EmailTypeTicketResolved:
		return r.resolveTicketResolvedRecipients(ctx, emailContext)
	case worker.EmailTypeTicketClosed:
		return r.resolveTicketClosedRecipients(ctx, emailContext)
	case worker.EmailTypeMessageReceived:
		return r.resolveMessageReceivedRecipients(ctx, emailContext)
	case worker.EmailTypeUserCreated:
		return r.resolveUserCreatedRecipients(ctx, emailContext)
	case worker.EmailTypePasswordReset:
		return r.resolvePasswordResetRecipients(ctx, emailContext)
	case worker.EmailTypeSystemMaintenance:
		return r.resolveSystemMaintenanceRecipients(ctx, emailContext)
	default:
		return nil, fmt.Errorf("未知的邮件类型: %s", emailType)
	}
}

// resolveTicketCreatedRecipients 工单创建时的收件人（通知管理员）
func (r *DefaultRecipientResolver) resolveTicketCreatedRecipients(ctx context.Context, emailContext map[string]interface{}) ([]string, error) {
	// 简化设计：所有工单都通知同一个管理员邮箱
	// 如果需要更复杂的分配逻辑，可以自定义RecipientResolver
	return []string{r.defaultAdminEmail}, nil
}

// resolveTicketClaimedRecipients 工单被接单时的收件人（通知创建者）
func (r *DefaultRecipientResolver) resolveTicketClaimedRecipients(ctx context.Context, emailContext map[string]interface{}) ([]string, error) {
	// 通知工单创建者
	if creatorEmail, ok := emailContext["creator_email"].(string); ok {
		return []string{creatorEmail}, nil
	}
	return nil, fmt.Errorf("工单创建者邮箱信息缺失")
}

// resolveTicketResolvedRecipients 工单已处理时的收件人
func (r *DefaultRecipientResolver) resolveTicketResolvedRecipients(ctx context.Context, emailContext map[string]interface{}) ([]string, error) {
	var recipients []string

	// 通知工单创建者
	if creatorEmail, ok := emailContext["creator_email"].(string); ok {
		recipients = append(recipients, creatorEmail)
	}

	// 可以同时通知处理者
	if handlerEmail, ok := emailContext["handler_email"].(string); ok {
		recipients = append(recipients, handlerEmail)
	}

	if len(recipients) == 0 {
		return nil, fmt.Errorf("没有找到有效的收件人")
	}

	return recipients, nil
}

// resolveTicketClosedRecipients 工单关闭时的收件人
func (r *DefaultRecipientResolver) resolveTicketClosedRecipients(ctx context.Context, emailContext map[string]interface{}) ([]string, error) {
	// 类似于工单已处理的逻辑
	return r.resolveTicketResolvedRecipients(ctx, emailContext)
}

// resolveMessageReceivedRecipients 收到新消息时的收件人
func (r *DefaultRecipientResolver) resolveMessageReceivedRecipients(ctx context.Context, emailContext map[string]interface{}) ([]string, error) {
	// 通知工单的相关人员（创建者和处理者）
	var recipients []string

	if creatorEmail, ok := emailContext["creator_email"].(string); ok {
		recipients = append(recipients, creatorEmail)
	}

	if handlerEmail, ok := emailContext["handler_email"].(string); ok && handlerEmail != emailContext["creator_email"] {
		recipients = append(recipients, handlerEmail)
	}

	return recipients, nil
}

// resolveUserCreatedRecipients 新用户创建时的收件人（通知管理员）
func (r *DefaultRecipientResolver) resolveUserCreatedRecipients(ctx context.Context, emailContext map[string]interface{}) ([]string, error) {
	// 通知管理员有新用户注册
	return []string{r.defaultAdminEmail}, nil
}

// resolvePasswordResetRecipients 密码重置时的收件人（通知用户本人）
func (r *DefaultRecipientResolver) resolvePasswordResetRecipients(ctx context.Context, emailContext map[string]interface{}) ([]string, error) {
	// 通知用户本人
	if userEmail, ok := emailContext["user_email"].(string); ok {
		return []string{userEmail}, nil
	}
	return nil, fmt.Errorf("用户邮箱信息缺失")
}

// resolveSystemMaintenanceRecipients 系统维护时的收件人
func (r *DefaultRecipientResolver) resolveSystemMaintenanceRecipients(ctx context.Context, emailContext map[string]interface{}) ([]string, error) {
	// 系统维护统一通知管理员
	return []string{r.defaultAdminEmail}, nil
}
