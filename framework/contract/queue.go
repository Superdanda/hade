package contract

import "context"

const QueueKey = "hade:queue"

type QueueService interface {
	// PublishEvent 发布事件到队列，并持久化
	PublishEvent(ctx context.Context, event Event) error

	// SubscribeEvent 订阅事件
	SubscribeEvent(ctx context.Context, topic string, handler func(event Event) error) error

	// ReplayEvents 从指定的时间点或事件ID开始重放事件
	ReplayEvents(ctx context.Context, topic string, fromID string, fromTimestamp int64, handler func(event Event) error) error

	// GetEventById 根据事件ID获取事件
	GetEventById(ctx context.Context, topic string, eventID string) (Event, error)

	// GetEventByTime 根据事件ID获取事件
	GetEventByTime(ctx context.Context, topic string, fromTimestamp int64) (Event, error)

	// Close 关闭队列连接
	Close() error

	// NewEventAndPublish 创建并推送事件方法
	NewEventAndPublish(ctx context.Context, topic string, payload interface{}) error
}

type Event interface {
	GetEventKey() string   // 事件唯一标识
	EventTopic() string    // 事件类型
	EventTimestamp() int64 // 事件发生时间
	Payload() interface{}  // 事件负载
}
