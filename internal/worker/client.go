package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// Client 异步任务客户端
type Client struct {
	client *asynq.Client
}

// NewClient 创建新的任务客户端
func NewClient(redisAddr string) *Client {
	return &Client{
		client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr}),
	}
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	return c.client.Close()
}

// EnqueueEmailTask 将邮件任务加入队列
func (c *Client) EnqueueEmailTask(ctx context.Context, task *EmailTask, opts ...asynq.Option) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("序列化邮件任务失败: %w", err)
	}

	taskInfo := asynq.NewTask(TypeEmailNotification, payload, opts...)
	_, err = c.client.EnqueueContext(ctx, taskInfo)
	if err != nil {
		return fmt.Errorf("入队邮件任务失败: %w", err)
	}

	return nil
}

// EnqueueEmailTaskIn 延迟发送邮件任务
func (c *Client) EnqueueEmailTaskIn(ctx context.Context, task *EmailTask, delay time.Duration) error {
	return c.EnqueueEmailTask(ctx, task, asynq.ProcessIn(delay))
}

// EnqueueEmailTaskAt 定时发送邮件任务
func (c *Client) EnqueueEmailTaskAt(ctx context.Context, task *EmailTask, t time.Time) error {
	return c.EnqueueEmailTask(ctx, task, asynq.ProcessAt(t))
}
