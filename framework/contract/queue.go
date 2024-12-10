package contract

import (
	"context"
	"encoding/json"
	"fmt"
)

const QueueKey = "hade:queue"

type QueueService interface {
	// PublishEvent 发布事件到队列，并持久化
	PublishEvent(ctx context.Context, event Event) error

	// SubscribeEvent 订阅事件
	SubscribeEvent(ctx context.Context, topic string, handler func(event Event) error) error

	// ProcessSubscribe 执行订阅
	ProcessSubscribe()

	// RegisterSubscribe 注册订阅 订阅事件
	RegisterSubscribe(topic string, handler func(event Event) error) error

	// ReplayEvents 从指定的时间点或事件ID开始重放事件
	ReplayEvents(ctx context.Context, topic string, fromID string, fromTimestamp int64, handler func(event Event) error) error

	// GetRegisterSubscribe 根据主题获取已经注册的订阅
	GetRegisterSubscribe(topic string) []EventHandler

	// GetEventById 根据事件ID获取事件
	GetEventById(ctx context.Context, topic string, eventID string) (Event, error)

	// GetEventByTime 根据事件ID获取事件
	GetEventByTime(ctx context.Context, topic string, fromTimestamp int64) (Event, error)

	// Close 关闭队列连接
	Close() error

	// NewEventAndPublish 创建并推送事件方法
	NewEventAndPublish(ctx context.Context, topic string, payload interface{}) error

	// SetContext 为订阅设置上下文
	SetContext(ctx context.Context)

	// GetContext 为订阅设置上下文
	GetContext() context.Context
}

type EventHandler func(event Event) error

type Event interface {
	GetEventKey() string       // 事件唯一标识
	EventTopic() string        // 事件类型
	EventTimestamp() int64     // 事件发生时间
	EventPayload() interface{} // 事件负载
	EventSource() string
}

// GetPayload 泛型方法，用于解析 Payload
func GetPayload[T any](event Event) (T, error) {
	// 定义一个零值变量
	var result T
	// 将 Payload 转为 JSON 字符串
	payload := event.EventPayload()
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return result, fmt.Errorf("failed to marshal payload: %w", err)
	}
	// 将 JSON 字符串解析为目标类型
	err = json.Unmarshal(payloadBytes, &result)
	if err != nil {
		return result, fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	return result, nil
}
