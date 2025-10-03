package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net/smtp"
	"strings"

	"student-services-platform-backend/internal/worker"
)

// Config 邮件服务配置
type Config struct {
	SMTPHost      string `mapstructure:"smtp_host"`      // SMTP服务器地址
	SMTPPort      int    `mapstructure:"smtp_port"`      // SMTP端口
	SMTPUsername  string `mapstructure:"smtp_username"`  // SMTP用户名
	SMTPPassword  string `mapstructure:"smtp_password"`  // SMTP密码
	FromEmail     string `mapstructure:"from_email"`     // 发件人邮箱
	FromName      string `mapstructure:"from_name"`      // 发件人姓名
	TLSEnabled    bool   `mapstructure:"tls_enabled"`    // 是否启用TLS
	TemplatesPath string `mapstructure:"templates_path"` // 模板文件路径
}

// Service 邮件服务
type Service struct {
	config            *Config
	recipientResolver RecipientResolver
	templateEngine    *TemplateEngine
}

// NewService 创建邮件服务
func NewService(config *Config) (*Service, error) {
	// 创建模板引擎
	templateEngine, err := NewTemplateEngine(config.TemplatesPath)
	if err != nil {
		return nil, fmt.Errorf("创建模板引擎失败: %w", err)
	}

	return &Service{
		config:            config,
		recipientResolver: NewDefaultRecipientResolver(config.FromEmail), // 传入发件人邮箱作为默认管理员邮箱
		templateEngine:    templateEngine,
	}, nil
}

// NewServiceWithResolver 创建邮件服务（自定义收件人解析器）
func NewServiceWithResolver(config *Config, resolver RecipientResolver) (*Service, error) {
	// 创建模板引擎
	templateEngine, err := NewTemplateEngine(config.TemplatesPath)
	if err != nil {
		return nil, fmt.Errorf("创建模板引擎失败: %w", err)
	}

	return &Service{
		config:            config,
		recipientResolver: resolver,
		templateEngine:    templateEngine,
	}, nil
}

// SendEmail 发送邮件
func (s *Service) SendEmail(ctx context.Context, task *worker.EmailTask) error {
	// 编码中文标题
	encodedSubject := mime.QEncoding.Encode("UTF-8", task.Subject)

	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)
	headers["To"] = strings.Join(task.To, ", ")
	headers["Subject"] = encodedSubject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// 构建邮件内容（不使用Base64编码）
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + task.Body

	// 发送邮件
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	// 根据端口和TLS配置选择发送方式
	if s.config.SMTPPort == 465 || !s.config.TLSEnabled {
		// 使用SSL连接（465端口）或不使用TLS
		err := s.sendMailSSL(addr, s.config.SMTPUsername, s.config.SMTPPassword, s.config.FromEmail, task.To, []byte(message))
		if err != nil {
			return fmt.Errorf("SMTP发送失败: %w", err)
		}
	} else {
		// 使用STARTTLS（587端口）
		auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
		err := smtp.SendMail(addr, auth, s.config.FromEmail, task.To, []byte(message))
		if err != nil {
			return fmt.Errorf("SMTP发送失败: %w", err)
		}
	}

	return nil
}

// sendMailSSL 使用SSL连接发送邮件（适用于465端口）
func (s *Service) sendMailSSL(addr, username, password, from string, to []string, msg []byte) error {
	// 创建TLS连接
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.config.SMTPHost,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS连接失败: %w", err)
	}
	defer conn.Close()

	// 创建SMTP客户端
	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("创建SMTP客户端失败: %w", err)
	}
	defer client.Close()

	// 身份验证
	auth := smtp.PlainAuth("", username, password, s.config.SMTPHost)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP身份验证失败: %w", err)
	}

	// 设置发件人
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}

	// 设置收件人
	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("设置收件人失败: %w", err)
		}
	}

	// 发送邮件内容
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("开始发送邮件数据失败: %w", err)
	}

	_, err = wc.Write(msg)
	if err != nil {
		return fmt.Errorf("写入邮件数据失败: %w", err)
	}

	err = wc.Close()
	if err != nil {
		return fmt.Errorf("关闭邮件数据流失败: %w", err)
	}

	return nil
}

// SendTemplateEmail 发送模板邮件（符合worker接口要求）
func (s *Service) SendTemplateEmail(ctx context.Context, emailType string, context map[string]interface{}) error {
	// 渲染模板
	templateBody, err := s.templateEngine.Render(emailType, context)
	if err != nil {
		return fmt.Errorf("模板渲染失败: %w", err)
	}

	// 解析收件人
	recipients, err := s.recipientResolver.ResolveRecipients(ctx, worker.EmailType(emailType), context)
	if err != nil {
		return fmt.Errorf("解析收件人失败: %w", err)
	}

	if len(recipients) == 0 {
		return fmt.Errorf("没有找到有效的收件人")
	}

	// 创建邮件任务
	task := &worker.EmailTask{
		To:       recipients,
		Subject:  s.getSubjectFromContext(context),
		Body:     templateBody,
		Type:     worker.EmailType(emailType),
		Priority: worker.EmailPriorityNormal,
		Context:  context,
	}

	// 发送邮件
	return s.SendEmail(ctx, task)
}

// getSubjectFromContext 从上下文中提取主题
func (s *Service) getSubjectFromContext(context map[string]interface{}) string {
	if title, ok := context["title"]; ok {
		if titleStr, ok := title.(string); ok {
			return titleStr
		}
	}
	return "工单通知"
}

// SendEmailWithDynamicRecipients 发送邮件（动态确定收件人）
func (s *Service) SendEmailWithDynamicRecipients(ctx context.Context, emailType worker.EmailType, subject, body string, emailContext map[string]interface{}) error {
	// 解析收件人
	recipients, err := s.recipientResolver.ResolveRecipients(ctx, emailType, emailContext)
	if err != nil {
		return fmt.Errorf("解析收件人失败: %w", err)
	}

	if len(recipients) == 0 {
		return fmt.Errorf("没有找到有效的收件人")
	}

	// 如果 body 为空，尝试使用模板渲染
	if body == "" {
		templateBody, err := s.templateEngine.Render(string(emailType), emailContext)
		if err != nil {
			return fmt.Errorf("模板渲染失败: %w", err)
		}
		body = templateBody
	}

	// 创建邮件任务
	task := &worker.EmailTask{
		To:       recipients,
		Subject:  subject,
		Body:     body,
		Type:     emailType,
		Priority: worker.EmailPriorityNormal,
		Context:  emailContext,
	}

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
