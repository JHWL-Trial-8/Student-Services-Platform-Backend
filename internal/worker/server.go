package worker

import (
	"log"
	"time"

	"github.com/hibiken/asynq"
)

// Server 异步任务服务器
type Server struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

// NewServer 创建新的任务服务器
func NewServer(redisAddr string, concurrency int) *Server {
	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: concurrency,
			Queues: map[string]int{
				"critical": 6, // 紧急队列
				"default":  3, // 默认队列
				"low":      1, // 低优先级队列
			},
			// 错误重试策略 - 不重试，一次失败直接放弃
			RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
				// 不重试，直接失败
				log.Printf("任务 %s 失败，不重试：%v", task.Type(), err)
				return 0 // 返回0表示不重试
			},
			// 日志配置
			Logger: nil, // 使用 asynq 默认日志器
		},
	)

	mux := asynq.NewServeMux()

	return &Server{
		server: server,
		mux:    mux,
	}
}

// RegisterHandlers 注册任务处理器
func (s *Server) RegisterHandlers(emailService EmailService) {
	emailHandler := NewEmailHandler(emailService)

	// 注册邮件发送任务处理器
	s.mux.HandleFunc(TypeEmailNotification, emailHandler.HandleEmailTask)
}

// Start 启动任务服务器
func (s *Server) Start() error {
	log.Println("Worker server starting...")
	return s.server.Start(s.mux)
}

// Stop 停止任务服务器
func (s *Server) Stop() {
	log.Println("Worker server stopping...")
	s.server.Stop()
}

// Shutdown 优雅关闭任务服务器
func (s *Server) Shutdown() {
	log.Println("Worker server shutting down...")
	s.server.Shutdown()
}
