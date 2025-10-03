# 邮件模板系统

## 概述

本项目采用基于文件的邮件模板系统，实现了前后端分离的设计理念。后端负责业务逻辑和数据处理，前端（模板文件）负责邮件的样式和布局。

## 设计原则

1. **前后端分离**：HTML模板由前端维护，后端只负责数据传递
2. **配置化管理**：通过配置文件指定模板目录位置
3. **热更新支持**：开发环境支持模板热重载
4. **类型安全**：模板渲染错误会被及时发现和处理

## 目录结构

```
templates/
└── email/
    ├── ticket_created.html      # 新工单创建通知
    ├── ticket_claimed.html      # 工单接单通知
    ├── ticket_resolved.html     # 工单处理完成通知
    ├── message_received.html    # 新消息通知
    └── ...                      # 其他模板
```

## 配置

在 `config.yaml` 中配置模板路径：

```yaml
email:
  # 其他邮件配置...
  templates_path: "./templates/email"  # 邮件模板文件夹路径
```

## 模板变量

每个模板都可以使用Go模板语法来渲染动态内容。常用变量包括：

### 工单相关模板变量
- `{{.ticket_id}}` - 工单ID
- `{{.title}}` - 工单标题
- `{{.category}}` - 工单分类
- `{{.is_urgent}}` - 是否紧急
- `{{.student_name}}` - 学生姓名
- `{{.student_email}}` - 学生邮箱
- `{{.admin_name}}` - 管理员姓名
- `{{.created_at}}` - 创建时间
- `{{.ticket_url}}` - 工单详情链接

### 消息相关模板变量
- `{{.recipient_name}}` - 收件人姓名
- `{{.sender_name}}` - 发件人姓名
- `{{.message_body}}` - 消息内容
- `{{.message_time}}` - 消息时间

## 使用方法

### 1. 在代码中使用模板引擎

```go
// 创建模板引擎
engine, err := email.NewTemplateEngine(config.Email.TemplatesPath)
if err != nil {
    log.Fatal(err)
}

// 渲染模板
data := map[string]interface{}{
    "ticket_id": "T123456",
    "title": "账号登录问题",
    "student_name": "张三",
    // ... 其他数据
}

htmlBody, err := engine.Render("ticket_created", data)
if err != nil {
    log.Printf("模板渲染失败: %v", err)
    return
}
```

### 2. 创建新模板

1. 在 `templates/email/` 目录下创建 `.html` 文件
2. 使用Go模板语法编写HTML内容
3. 模板名称为文件名（不含`.html`扩展名）

### 3. 模板热更新（开发环境）

```go
// 重新加载所有模板
err := engine.ReloadTemplates()
if err != nil {
    log.Printf("重载模板失败: %v", err)
}
```

## 模板开发指南

### HTML结构建议

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>邮件标题</title>
    <style>
        /* 内联CSS样式 */
        body { font-family: Arial, sans-serif; }
        .container { max-width: 600px; margin: 0 auto; }
        /* ... */
    </style>
</head>
<body>
    <div class="container">
        <!-- 邮件内容 -->
    </div>
</body>
</html>
```

### 样式建议

1. **使用内联CSS**：许多邮件客户端不支持外部CSS
2. **响应式设计**：考虑移动设备的显示效果
3. **兼容性**：避免使用新的CSS特性
4. **颜色主题**：保持与系统一致的品牌色彩

### 模板语法示例

```html
<!-- 条件判断 -->
{{if .is_urgent}}
    <span style="color: red;">🔴 紧急</span>
{{else}}
    <span style="color: green;">🟢 普通</span>
{{end}}

<!-- 循环遍历 -->
{{range .items}}
    <li>{{.name}}: {{.value}}</li>
{{end}}

<!-- 变量输出 -->
<p>工单编号：{{.ticket_id}}</p>
<p>创建时间：{{.created_at}}</p>
```

## 错误处理

系统会自动处理以下错误情况：

1. **模板文件不存在**：返回明确的错误信息
2. **模板语法错误**：在加载时检测并报告
3. **变量不存在**：渲染时安全处理，不会导致崩溃
4. **目录权限问题**：启动时检查并报告

## 性能优化

1. **模板缓存**：模板加载后会缓存在内存中
2. **按需加载**：只加载需要的模板文件
3. **错误缓存**：避免重复尝试加载失败的模板

## 最佳实践

1. **版本控制**：将模板文件纳入Git版本控制
2. **测试验证**：使用预览功能测试模板效果
3. **文档维护**：及时更新模板变量说明
4. **备份策略**：重要模板文件要有备份机制

## 扩展功能

未来可以考虑添加：

1. **多语言支持**：根据用户语言选择不同模板
2. **主题切换**：支持多套视觉主题
3. **模板继承**：支持模板布局继承机制
4. **在线编辑**：提供Web界面编辑模板