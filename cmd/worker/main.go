package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"student-services-platform-backend/internal/config"
	"student-services-platform-backend/internal/email"
	"student-services-platform-backend/internal/worker"
)

// 独立的worker进程示例

func main() {
	// 加载配置
	cfg := config.MustLoad()

	// 创建邮件服务
	emailConfig := &email.Config{
		SMTPHost:      cfg.Email.SMTPHost,
		SMTPPort:      cfg.Email.SMTPPort,
		SMTPUsername:  cfg.Email.SMTPUsername,
		SMTPPassword:  cfg.Email.SMTPPassword,
		FromEmail:     cfg.Email.FromEmail,
		FromName:      cfg.Email.FromName,
		TLSEnabled:    cfg.Email.TLSEnabled,
		TemplatesPath: cfg.Email.TemplatesPath,
	}

	emailService, err := email.NewService(emailConfig)
	if err != nil {
		log.Fatalf("创建邮件服务失败: %v", err)
	}

	// 验证邮件配置
	if err := emailService.ValidateConfig(); err != nil {
		log.Fatalf("邮件配置无效: %v", err)
	}

	// 创建Worker管理器
	workerManager := worker.NewManager(emailService)

	// 启动Worker服务器
	go func() {
		if err := workerManager.StartServer(); err != nil {
			log.Fatalf("启动Worker服务器失败: %v", err)
		}
	}()

	log.Println("邮件Worker服务已启动，等待任务...")

	// 等待系统信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("接收到退出信号，正在关闭...")

	// 优雅关闭
	workerManager.Shutdown()
	log.Println("Worker服务已关闭")
}
