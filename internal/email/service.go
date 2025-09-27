package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"student-services-platform-backend/internal/worker"
)

// Config 邮件服务配置
type Config struct {
	SMTPHost     string `mapstructure:"smtp_host"`     // SMTP服务器地址
	SMTPPort     int    `mapstructure:"smtp_port"`     // SMTP端口
	SMTPUsername string `mapstructure:"smtp_username"` // SMTP用户名
	SMTPPassword string `mapstructure:"smtp_password"` // SMTP密码
	FromEmail    string `mapstructure:"from_email"`    // 发件人邮箱
	FromName     string `mapstructure:"from_name"`     // 发件人姓名
	TLSEnabled   bool   `mapstructure:"tls_enabled"`   // 是否启用TLS
}

// Service 邮件服务
type Service struct {
	config         *Config
	templateEngine *TemplateEngine
}

// NewService 创建邮件服务
func NewService(config *Config) *Service {
	return &Service{
		config:         config,
		templateEngine: NewTemplateEngine(),
	}
}

// SendEmail 发送邮件
func (s *Service) SendEmail(ctx context.Context, task *worker.EmailTask) error {
	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)
	headers["To"] = strings.Join(task.To, ", ")
	headers["Subject"] = task.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// 构建邮件内容
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + task.Body

	// 发送邮件
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	err := smtp.SendMail(addr, auth, s.config.FromEmail, task.To, []byte(message))
	if err != nil {
		return fmt.Errorf("SMTP发送失败: %w", err)
	}

	return nil
}

// SendTemplateEmail 发送模板邮件
func (s *Service) SendTemplateEmail(ctx context.Context, task *worker.EmailTask) error {
	// 渲染邮件模板
	renderedBody, err := s.templateEngine.Render(string(task.Type), task.Context)
	if err != nil {
		return fmt.Errorf("渲染邮件模板失败: %w", err)
	}

	// 更新任务的邮件内容
	task.Body = renderedBody

	// 发送邮件
	return s.SendEmail(ctx, task)
}

// ValidateConfig 验证邮件配置
func (s *Service) ValidateConfig() error {
	if s.config.SMTPHost == "" {
		return fmt.Errorf("SMTP主机地址不能为空")
	}
	if s.config.SMTPPort <= 0 {
		return fmt.Errorf("SMTP端口必须大于0")
	}
	if s.config.FromEmail == "" {
		return fmt.Errorf("发件人邮箱不能为空")
	}
	return nil
}
