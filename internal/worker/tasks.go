package worker

import "time"

// 任务类型常量
const (
	TypeEmailNotification = "email:notification"
)

// EmailTask 邮件发送任务
type EmailTask struct {
	// 收件人信息
	To      []string `json:"to"`       // 收件人邮箱列表
	ToNames []string `json:"to_names"` // 收件人姓名列表

	// 邮件内容
	Subject  string `json:"subject"`   // 邮件主题
	Body     string `json:"body"`      // 邮件正文（HTML格式）
	TextBody string `json:"text_body"` // 纯文本正文（可选）

	// 邮件类型和上下文
	Type    EmailType              `json:"type"`    // 邮件类型
	Context map[string]interface{} `json:"context"` // 模板上下文数据

	// 发送配置
	Priority EmailPriority `json:"priority"` // 优先级
	Retry    int           `json:"retry"`    // 重试次数

	// 业务关联
	TicketID  *uint `json:"ticket_id,omitempty"`  // 关联工单ID
	MessageID *uint `json:"message_id,omitempty"` // 关联消息ID
	UserID    *uint `json:"user_id,omitempty"`    // 关联用户ID

	// 时间戳
	CreatedAt time.Time `json:"created_at"` // 任务创建时间
}

// EmailType 邮件类型枚举
type EmailType string

const (
	EmailTypeTicketCreated     EmailType = "ticket_created"     // 工单创建通知
	EmailTypeTicketClaimed     EmailType = "ticket_claimed"     // 工单被接单通知
	EmailTypeTicketUnclaimed   EmailType = "ticket_unclaimed"   // 工单被撤销通知
	EmailTypeTicketResolved    EmailType = "ticket_resolved"    // 工单已处理通知
	EmailTypeTicketClosed      EmailType = "ticket_closed"      // 工单已关闭通知
	EmailTypeTicketRated       EmailType = "ticket_rated"       // 工单被评价通知
	EmailTypeMessageReceived   EmailType = "message_received"   // 收到新消息通知
	EmailTypeSpamFlagged       EmailType = "spam_flagged"       // 垃圾标记通知
	EmailTypeSpamReviewed      EmailType = "spam_reviewed"      // 垃圾审核结果通知
	EmailTypeUserCreated       EmailType = "user_created"       // 用户创建通知
	EmailTypePasswordReset     EmailType = "password_reset"     // 密码重置通知
	EmailTypeSystemMaintenance EmailType = "system_maintenance" // 系统维护通知
)

// EmailPriority 邮件优先级
type EmailPriority string

const (
	EmailPriorityCritical EmailPriority = "critical" // 紧急（系统故障、安全问题）
	EmailPriorityHigh     EmailPriority = "high"     // 高优先级（重要通知）
	EmailPriorityNormal   EmailPriority = "normal"   // 普通优先级（常规通知）
	EmailPriorityLow      EmailPriority = "low"      // 低优先级（营销邮件等）
)

// GetQueueName 根据优先级获取队列名称
func (p EmailPriority) GetQueueName() string {
	switch p {
	case EmailPriorityCritical, EmailPriorityHigh:
		return "critical"
	case EmailPriorityNormal:
		return "default"
	case EmailPriorityLow:
		return "low"
	default:
		return "default"
	}
}

// NewEmailTask 创建新的邮件任务
func NewEmailTask(emailType EmailType, to []string, subject, body string) *EmailTask {
	return &EmailTask{
		To:        to,
		Subject:   subject,
		Body:      body,
		Type:      emailType,
		Priority:  EmailPriorityNormal,
		Retry:     3,
		Context:   make(map[string]interface{}),
		CreatedAt: time.Now(),
	}
}

// SetPriority 设置邮件优先级
func (t *EmailTask) SetPriority(priority EmailPriority) *EmailTask {
	t.Priority = priority
	return t
}

// SetContext 设置模板上下文
func (t *EmailTask) SetContext(key string, value interface{}) *EmailTask {
	t.Context[key] = value
	return t
}

// SetTicketID 设置关联工单ID
func (t *EmailTask) SetTicketID(ticketID uint) *EmailTask {
	t.TicketID = &ticketID
	return t
}

// SetUserID 设置关联用户ID
func (t *EmailTask) SetUserID(userID uint) *EmailTask {
	t.UserID = &userID
	return t
}
