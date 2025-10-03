package worker

import (
	"log"
)

// Server 邮件任务服务器
type Server struct {
	emailHandler *EmailHandler
}

// NewServer 创建新的任务服务器
func NewServer(emailService EmailService) *Server {
	emailHandler := NewEmailHandler(emailService)

	return &Server{
		emailHandler: emailHandler,
	}
}

// RegisterHandlers 注册任务处理器
func (s *Server) RegisterHandlers(emailService EmailService) {
	// 在同步模式下，不需要注册处理器
	// 处理器已经在创建时设置
}

// Start 启动任务服务器
func (s *Server) Start() error {
	log.Println("邮件任务服务器已启动（同步模式）")
	// 在同步模式下，不需要启动服务器
	return nil
}

// Stop 停止任务服务器
func (s *Server) Stop() {
	log.Println("邮件任务服务器已停止")
	// 在同步模式下，不需要停止服务器
}

// Shutdown 优雅关闭任务服务器
func (s *Server) Shutdown() {
	log.Println("邮件任务服务器已关闭")
	// 在同步模式下，不需要关闭服务器
}
