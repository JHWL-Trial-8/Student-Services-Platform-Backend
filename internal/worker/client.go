package worker

import (
	"context"
	"time"
)

// Client 邮件任务客户端
type Client struct {
	emailHandler *EmailHandler
}

// NewClient 创建新的任务客户端
func NewClient(emailHandler *EmailHandler) *Client {
	return &Client{
		emailHandler: emailHandler,
	}
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	// 不需要关闭任何资源
	return nil
}

// EnqueueEmailTask 将邮件任务加入队列
func (c *Client) EnqueueEmailTask(ctx context.Context, task *EmailTask) error {
	// 直接处理邮件任务，而不是加入队列
	return c.emailHandler.HandleEmailTask(ctx, task)
}

// EnqueueEmailTaskIn 延迟发送邮件任务
func (c *Client) EnqueueEmailTaskIn(ctx context.Context, task *EmailTask, delay time.Duration) error {
	// 在实际应用中，这里可以实现延迟发送逻辑
	// 但为了简化，我们直接发送
	return c.emailHandler.HandleEmailTask(ctx, task)
}

// EnqueueEmailTaskAt 定时发送邮件任务
func (c *Client) EnqueueEmailTaskAt(ctx context.Context, task *EmailTask, t time.Time) error {
	// 在实际应用中，这里可以实现定时发送逻辑
	// 但为了简化，我们直接发送
	return c.emailHandler.HandleEmailTask(ctx, task)
}
