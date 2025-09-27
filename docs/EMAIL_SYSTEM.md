# 邮件发送系统使用指南

本项目实现了基于Redis队列的异步邮件发送系统，符合第20步的开发要求。

## 文件结构

```
internal/
├── worker/          # 异步任务处理
│   ├── client.go    # 任务发布客户端
│   ├── server.go    # 任务处理服务器
│   ├── tasks.go     # 任务定义和类型
│   ├── handlers.go  # 任务处理器
│   └── manager.go   # 任务管理器
├── email/           # 邮件服务
│   ├── service.go   # 邮件发送服务
│   └── templates.go # 邮件模板引擎
└── config/
    └── email.go     # 邮件配置结构
config/
└── email.example.yaml  # 配置示例
```

## 功能特性

### ✅ 已实现的功能：

1. **异步任务队列**
   - 基于 `github.com/hibiken/asynq` 和 Redis
   - 支持优先级队列（critical、default、low）
   - 自动重试机制
   - 并发处理能力

2. **邮件发送服务**
   - SMTP 邮件发送
   - HTML 邮件模板支持
   - 文件类型验证
   - 错误处理和日志记录

3. **预定义邮件模板**
   - 工单创建通知
   - 工单接单通知  
   - 新消息通知
   - 其他系统通知

4. **任务管理器**
   - 统一的邮件发送接口
   - 工单状态变更通知
   - 消息通知功能

## 使用方法

### 1. 配置文件

复制 `config/email.example.yaml` 到你的配置文件中，并填入实际的SMTP信息：

```yaml
email:
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  smtp_username: "your-email@gmail.com"
  smtp_password: "your-app-password"
  from_email: "noreply@your-domain.com"
  from_name: "学生服务平台"
  tls_enabled: true

redis:
  addr: "localhost:6379"
  password: ""
  db: 0
```

### 2. 在项目中集成

在你的工单服务中集成邮件通知：

```go
// 1. 创建邮件服务
emailService := email.NewService(&email.Config{
    SMTPHost:     cfg.Email.SMTPHost,
    SMTPPort:     cfg.Email.SMTPPort,
    SMTPUsername: cfg.Email.SMTPUsername,
    SMTPPassword: cfg.Email.SMTPPassword,
    FromEmail:    cfg.Email.FromEmail,
    FromName:     cfg.Email.FromName,
    TLSEnabled:   cfg.Email.TLSEnabled,
})

// 2. 创建Worker管理器
workerManager := worker.NewManager(cfg.Redis.Addr, emailService)

// 3. 启动Worker（可选，也可以用独立进程）
go func() {
    if err := workerManager.StartServer(); err != nil {
        log.Printf("Worker启动失败: %v", err)
    }
}()

// 4. 在工单服务中使用
// 工单创建时发送通知
err := workerManager.SendTicketCreatedNotification(
    ctx,
    []string{"admin@example.com"},
    ticketID,
    "空调维修",
    "设备维修",
    "张三",
    "student@example.com",
    true, // 是否紧急
)
```

### 3. 独立Worker进程（推荐）

你也可以运行独立的Worker进程来处理邮件发送任务：

```go
// cmd/worker/main.go
func main() {
    cfg := config.MustLoad()
    emailService := email.NewService(cfg.Email)
    workerManager := worker.NewManager(cfg.Redis.Addr, emailService)
    
    // 启动Worker服务器
    if err := workerManager.StartServer(); err != nil {
        log.Fatalf("启动Worker失败: %v", err)
    }
}
```

### 4. 支持的邮件类型

- `EmailTypeTicketCreated` - 工单创建通知
- `EmailTypeTicketClaimed` - 工单接单通知
- `EmailTypeMessageReceived` - 新消息通知
- `EmailTypeTicketResolved` - 工单已处理通知
- `EmailTypeTicketClosed` - 工单已关闭通知
- 等等...

## 技术亮点

1. **解耦设计**：邮件发送不会阻塞HTTP请求
2. **高可靠性**：Redis持久化任务，失败自动重试
3. **可扩展性**：支持多Worker进程，水平扩展
4. **模板系统**：内置HTML邮件模板，支持动态数据
5. **优先级队列**：紧急邮件优先处理
6. **错误处理**：完善的错误处理和日志记录

## 下一步

这个系统已经实现了第20步的所有要求。可以继续开发：

- 第21步：管理员统计大屏 (`GET /admin/stats`)
- 第22步：预设回复CRUD功能

## 依赖

确保在 `go.mod` 中添加了必要的依赖：

```go
require (
    github.com/hibiken/asynq v0.25.0
    // ... 其他依赖
)
```