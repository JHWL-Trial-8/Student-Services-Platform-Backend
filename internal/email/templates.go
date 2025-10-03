package email

import (
	"bytes"
	"fmt"
	"html/template"
)

// TemplateEngine 邮件模板引擎
type TemplateEngine struct {
	templates map[string]*template.Template
}

// NewTemplateEngine 创建模板引擎
func NewTemplateEngine() *TemplateEngine {
	engine := &TemplateEngine{
		templates: make(map[string]*template.Template),
	}

	// 加载内置模板
	engine.loadBuiltinTemplates()

	return engine
}

// Render 渲染模板
func (e *TemplateEngine) Render(templateName string, data map[string]interface{}) (string, error) {
	tmpl, exists := e.templates[templateName]
	if !exists {
		return "", fmt.Errorf("模板不存在: %s", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("模板渲染失败: %w", err)
	}

	return buf.String(), nil
}

// loadBuiltinTemplates 加载内置邮件模板
func (e *TemplateEngine) loadBuiltinTemplates() {
	templates := map[string]string{
		"ticket_created": `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>新工单通知</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2c5aa0;">新工单创建通知</h2>
        <p>尊敬的管理员，</p>
        <p>有新的工单需要处理：</p>
        
        <div style="background: #f5f5f5; padding: 15px; border-radius: 5px; margin: 20px 0;">
            <p><strong>工单编号：</strong>{{.ticket_id}}</p>
            <p><strong>标题：</strong>{{.title}}</p>
            <p><strong>分类：</strong>{{.category}}</p>
            <p><strong>紧急程度：</strong>{{if .is_urgent}}紧急{{else}}普通{{end}}</p>
            <p><strong>提交人：</strong>{{.student_name}} ({{.student_email}})</p>
            <p><strong>创建时间：</strong>{{.created_at}}</p>
        </div>
        
        <p><strong>工单描述：</strong></p>
        <div style="background: #fff; border-left: 4px solid #2c5aa0; padding: 15px; margin: 20px 0;">
            {{.description}}
        </div>
        
        <p>
            <a href="{{.ticket_url}}" style="background: #2c5aa0; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">查看工单详情</a>
        </p>
        
        <hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">
        <p style="font-size: 12px; color: #666;">
            此邮件由学生服务平台自动发送，请勿直接回复。
        </p>
    </div>
</body>
</html>`,

		"ticket_claimed": `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>工单处理通知</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #28a745;">工单已被接单</h2>
        <p>{{.student_name}} 同学，</p>
        <p>您的工单已被管理员接单处理：</p>
        
        <div style="background: #f5f5f5; padding: 15px; border-radius: 5px; margin: 20px 0;">
            <p><strong>工单编号：</strong>{{.ticket_id}}</p>
            <p><strong>标题：</strong>{{.title}}</p>
            <p><strong>处理人员：</strong>{{.admin_name}}</p>
            <p><strong>接单时间：</strong>{{.claimed_at}}</p>
        </div>
        
        <p>管理员将尽快为您处理工单，如有疑问可通过工单消息功能进行沟通。</p>
        
        <p>
            <a href="{{.ticket_url}}" style="background: #28a745; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">查看工单详情</a>
        </p>
        
        <hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">
        <p style="font-size: 12px; color: #666;">
            此邮件由学生服务平台自动发送，请勿直接回复。
        </p>
    </div>
</body>
</html>`,

		"ticket_resolved": `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>工单处理完成通知</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #28a745;">🎉 您的工单已处理完成</h2>
        <p>{{.student_name}} 同学，</p>
        <p>您提交的工单已经处理完成：</p>
        
        <div style="background: #f5f5f5; padding: 15px; border-radius: 5px; margin: 20px 0;">
            <p><strong>工单编号：</strong>{{.ticket_id}}</p>
            <p><strong>标题：</strong>{{.title}}</p>
            <p><strong>处理人员：</strong>{{.admin_name}}</p>
            <p><strong>处理时间：</strong>{{.resolved_at}}</p>
        </div>
        
        <div style="background: #e8f5e8; padding: 15px; border-radius: 5px; margin: 20px 0; border-left: 4px solid #28a745;">
            <h3 style="margin-top: 0; color: #155724;">处理结果：</h3>
            <p style="margin-bottom: 0;">{{.resolution}}</p>
        </div>
        
        <p>如果您对处理结果满意，欢迎给我们评价。如有任何问题，请随时联系我们。</p>
        
        <p>
            <a href="{{.ticket_url}}" style="background: #28a745; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">查看工单详情</a>
        </p>
        
        <hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">
        <p style="font-size: 12px; color: #666;">
            此邮件由学生服务平台自动发送，请勿直接回复。
        </p>
    </div>
</body>
</html>`,

		"message_received": `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>新消息通知</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #007bff;">新消息通知</h2>
        <p>{{.recipient_name}}，</p>
        <p>您的工单收到了新的回复：</p>
        
        <div style="background: #f5f5f5; padding: 15px; border-radius: 5px; margin: 20px 0;">
            <p><strong>工单编号：</strong>{{.ticket_id}}</p>
            <p><strong>标题：</strong>{{.title}}</p>
            <p><strong>回复人：</strong>{{.sender_name}}</p>
            <p><strong>回复时间：</strong>{{.message_time}}</p>
        </div>
        
        <p><strong>消息内容：</strong></p>
        <div style="background: #fff; border-left: 4px solid #007bff; padding: 15px; margin: 20px 0;">
            {{.message_body}}
        </div>
        
        <p>
            <a href="{{.ticket_url}}" style="background: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">查看并回复</a>
        </p>
        
        <hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">
        <p style="font-size: 12px; color: #666;">
            此邮件由学生服务平台自动发送，请勿直接回复。
        </p>
    </div>
</body>
</html>`,
	}

	for name, content := range templates {
		tmpl, err := template.New(name).Parse(content)
		if err != nil {
			// 在实际项目中应该记录日志而不是panic
			panic(fmt.Sprintf("加载模板失败 %s: %v", name, err))
		}
		e.templates[name] = tmpl
	}
}

// RegisterTemplate 注册自定义模板
func (e *TemplateEngine) RegisterTemplate(name, content string) error {
	tmpl, err := template.New(name).Parse(content)
	if err != nil {
		return fmt.Errorf("解析模板失败: %w", err)
	}

	e.templates[name] = tmpl
	return nil
}

// ListTemplates 列出所有可用模板
func (e *TemplateEngine) ListTemplates() []string {
	var names []string
	for name := range e.templates {
		names = append(names, name)
	}
	return names
}
