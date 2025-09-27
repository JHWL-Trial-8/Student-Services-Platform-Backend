package worker

import (
	"context"
	"fmt"
	"log"
)

// Manager 异步任务管理器
type Manager struct {
	client       *Client
	server       *Server
	emailService EmailService
}

// NewManager 创建任务管理器
func NewManager(redisAddr string, emailService EmailService) *Manager {
	client := NewClient(redisAddr)
	server := NewServer(redisAddr, 10) // 10个并发worker

	// 注册任务处理器
	server.RegisterHandlers(emailService)

	return &Manager{
		client:       client,
		server:       server,
		emailService: emailService,
	}
}

// GetClient 获取任务客户端
func (m *Manager) GetClient() *Client {
	return m.client
}

// StartServer 启动任务服务器
func (m *Manager) StartServer() error {
	log.Println("启动异步任务处理服务器...")
	return m.server.Start()
}

// StopServer 停止任务服务器
func (m *Manager) StopServer() {
	m.server.Stop()
}

// Shutdown 关闭
func (m *Manager) Shutdown() {
	log.Println("关闭异步任务管理器...")
	m.server.Shutdown()
	if err := m.client.Close(); err != nil {
		log.Printf("关闭任务客户端失败: %v", err)
	}
}

// SendTicketCreatedNotification 发送工单创建通知
func (m *Manager) SendTicketCreatedNotification(ctx context.Context,
	adminEmails []string, ticketID uint, title, category, studentName, studentEmail string, isUrgent bool) error {

	task := NewEmailTask(EmailTypeTicketCreated, adminEmails, "新工单需要处理", "").
		SetPriority(func() EmailPriority {
			if isUrgent {
				return EmailPriorityHigh
			}
			return EmailPriorityNormal
		}()).
		SetTicketID(ticketID).
		SetContext("ticket_id", ticketID).
		SetContext("title", title).
		SetContext("category", category).
		SetContext("student_name", studentName).
		SetContext("student_email", studentEmail).
		SetContext("is_urgent", isUrgent).
		SetContext("ticket_url", getTicketURL(ticketID))

	return m.client.EnqueueEmailTask(ctx, task)
}

// SendTicketClaimedNotification 发送工单接单通知
func (m *Manager) SendTicketClaimedNotification(ctx context.Context,
	studentEmail, studentName, adminName string, ticketID uint, title string) error {

	task := NewEmailTask(EmailTypeTicketClaimed, []string{studentEmail}, "您的工单已被接单", "").
		SetPriority(EmailPriorityNormal).
		SetTicketID(ticketID).
		SetContext("ticket_id", ticketID).
		SetContext("title", title).
		SetContext("student_name", studentName).
		SetContext("admin_name", adminName).
		SetContext("ticket_url", getTicketURL(ticketID))

	return m.client.EnqueueEmailTask(ctx, task)
}

// SendMessageNotification 发送新消息通知
func (m *Manager) SendMessageNotification(ctx context.Context,
	recipientEmail, recipientName, senderName, messageBody string, ticketID uint, title string) error {

	task := NewEmailTask(EmailTypeMessageReceived, []string{recipientEmail}, "工单收到新回复", "").
		SetPriority(EmailPriorityNormal).
		SetTicketID(ticketID).
		SetContext("ticket_id", ticketID).
		SetContext("title", title).
		SetContext("recipient_name", recipientName).
		SetContext("sender_name", senderName).
		SetContext("message_body", messageBody).
		SetContext("ticket_url", getTicketURL(ticketID))

	return m.client.EnqueueEmailTask(ctx, task)
}

// getTicketURL 生成工单链接
func getTicketURL(ticketID uint) string {
	// 这里应该从配置中获取前端地址
	return fmt.Sprintf("https://your-frontend-domain.com/tickets/%d", ticketID)
}
