package email

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// TemplateEngine 邮件模板引擎
type TemplateEngine struct {
	templates     map[string]*template.Template
	templatesPath string
}

// NewTemplateEngine 创建模板引擎
func NewTemplateEngine(templatesPath string) (*TemplateEngine, error) {
	engine := &TemplateEngine{
		templates:     make(map[string]*template.Template),
		templatesPath: templatesPath,
	}

	// 加载外部模板文件
	if err := engine.loadTemplatesFromFiles(); err != nil {
		return nil, fmt.Errorf("加载模板文件失败: %w", err)
	}

	return engine, nil
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

// loadTemplatesFromFiles 从文件系统加载模板
func (e *TemplateEngine) loadTemplatesFromFiles() error {
	// 检查模板目录是否存在
	if e.templatesPath == "" {
		return fmt.Errorf("模板路径不能为空")
	}

	// 检查目录是否存在
	if _, err := os.Stat(e.templatesPath); os.IsNotExist(err) {
		return fmt.Errorf("模板目录不存在: %s", e.templatesPath)
	}

	// 遍历模板目录
	err := filepath.Walk(e.templatesPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理 .html 文件
		if !strings.HasSuffix(strings.ToLower(path), ".html") {
			return nil
		}

		// 获取模板名称（不包含扩展名）
		templateName := strings.TrimSuffix(filepath.Base(path), ".html")

		// 加载模板文件
		tmpl, err := template.ParseFiles(path)
		if err != nil {
			return fmt.Errorf("解析模板文件 %s 失败: %w", path, err)
		}

		// 存储模板
		e.templates[templateName] = tmpl

		return nil
	})

	if err != nil {
		return fmt.Errorf("遍历模板目录失败: %w", err)
	}

	return nil
}

// ReloadTemplates 重新加载模板（用于开发环境热更新）
func (e *TemplateEngine) ReloadTemplates() error {
	// 清空现有模板
	e.templates = make(map[string]*template.Template)

	// 重新加载
	return e.loadTemplatesFromFiles()
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

// HasTemplate 检查模板是否存在
func (e *TemplateEngine) HasTemplate(name string) bool {
	_, exists := e.templates[name]
	return exists
}

// GetTemplatesPath 获取模板路径
func (e *TemplateEngine) GetTemplatesPath() string {
	return e.templatesPath
}
