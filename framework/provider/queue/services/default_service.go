package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
	"github.com/Superdanda/hade/framework/provider/orm"
	"gorm.io/gorm"
	"strconv"
	"sync"
	"time"
)

type MemoryQueueService struct {
	container          framework.Container
	subscribers        map[string][]func(event contract.Event) error
	subscriberMu       sync.RWMutex
	eventStore         EventStore
	eventQueue         chan contract.Event // 使用 channel 作为 FIFO 队列
	RegisterSubscribed map[string][]contract.EventHandler
	context            context.Context
}

func (m *MemoryQueueService) SetContext(ctx context.Context) {
	m.context = ctx
}

func (m *MemoryQueueService) ProcessSubscribe() {

}

func (m *MemoryQueueService) RegisterSubscribe(topic string, handler func(event contract.Event) error) error {
	m.RegisterSubscribed[topic] = append(m.RegisterSubscribed[topic], handler)
	return nil
}

func (m *MemoryQueueService) GetRegisterSubscribe(topic string) []contract.EventHandler {
	return m.RegisterSubscribed[topic]
}

func (m *MemoryQueueService) NewEventAndPublish(ctx context.Context, topic string, payload interface{}) error {
	marshal, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	defaultBaseEvent := &DefaultBaseEvent{
		Topic:     topic,
		Timestamp: time.Now().UnixNano(),
		Data:      string(marshal),
	}

	err = m.PublishEvent(ctx, defaultBaseEvent)
	if err != nil {
		return err
	}

	err = m.eventStore.SaveEvent(ctx, defaultBaseEvent)
	if err != nil {
		return err
	}
	return nil
}

func NewMemoryQueueService(params ...interface{}) (interface{}, error) {
	container := params[0].(framework.Container)
	memoryQueueService := &MemoryQueueService{container: container,
		subscribers: make(map[string][]func(event contract.Event) error),
		eventQueue:  make(chan contract.Event, 100), // 定义队列长度
	}
	// 创建eventStore
	ormService := container.MustMake(contract.ORMKey).(contract.ORMService)
	db, err := ormService.GetDB(orm.WithConfigPath("database.default"))
	if err != nil {
		return nil, err
	}
	eventStore := NewGormEventStore(db)
	memoryQueueService.eventStore = eventStore

	go memoryQueueService.processEvents() // 启动事件处理协程

	memoryQueueService.RegisterSubscribed = make(map[string][]contract.EventHandler)
	return memoryQueueService, nil
}

func (m *MemoryQueueService) processEvents() {
	for event := range m.eventQueue {
		m.subscriberMu.RLock()
		handlers := m.subscribers[event.EventTopic()]
		m.subscriberMu.RUnlock()
		// 按顺序调用订阅者
		for _, handler := range handlers {
			go handler(event)
		}
		// 保存事件到持久化存储
		_ = m.eventStore.SaveEvent(context.Background(), event)
	}
}

func (m *MemoryQueueService) PublishEvent(ctx context.Context, event contract.Event) error {
	select {
	case m.eventQueue <- event:
		return nil
	default:
		return fmt.Errorf("event queue is full")
	}
}

func (m *MemoryQueueService) SubscribeEvent(ctx context.Context, eventType string, handler func(event contract.Event) error) error {
	m.subscriberMu.Lock()
	defer m.subscriberMu.Unlock()

	if m.subscribers == nil {
		m.subscribers = make(map[string][]func(event contract.Event) error)
	}

	m.subscribers[eventType] = append(m.subscribers[eventType], handler)
	return nil
}

func (m *MemoryQueueService) ReplayEvents(ctx context.Context, topic string, fromID string, fromTimestamp int64, handler func(event contract.Event) error) error {
	// 转换 fromID 为 int64
	var startID int64
	if fromID != "" {
		var err error
		startID, err = strconv.ParseInt(fromID, 10, 64)
		if err != nil {
			return err
		}
	}

	// 从 eventStore 获取事件
	events, err := m.eventStore.GetEvents(ctx, "", startID, fromTimestamp)
	if err != nil {
		return err
	}

	// 调用处理函数
	for _, event := range events {
		if err := handler(event); err != nil {
			return err
		}
	}
	return nil
}

func (m *MemoryQueueService) GetEventById(ctx context.Context, topic string, eventID string) (contract.Event, error) {
	id, err := strconv.ParseInt(eventID, 10, 64)
	if err != nil {
		return nil, err
	}
	return m.eventStore.GetEventByID(ctx, id)
}

func (m *MemoryQueueService) GetEventByTime(ctx context.Context, topic string, fromTimestamp int64) (contract.Event, error) {
	events, err := m.eventStore.GetEvents(ctx, topic, 0, fromTimestamp)
	if err != nil {
		return nil, err
	}
	if len(events) > 0 {
		return events[0], nil
	}
	return nil, fmt.Errorf("no events found")
}

func (m *MemoryQueueService) Close() error {
	// 内存队列不需要清理操作，因此可以返回 nil
	return nil
}

type DefaultBaseEvent struct {
	ID        int64     `gorm:"primaryKey;type:bigint" json:"id"`
	EventKey  string    `gorm:"type:varchar(255);not null" json:"event_key"`
	Topic     string    `gorm:"type:varchar(50);not null" json:"topic"`
	Timestamp int64     `gorm:"autoCreateTime:milli" json:"timestamp"`
	Data      string    `gorm:"type:json" json:"data"` // 将数据存储为 JSON 格式
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (d *DefaultBaseEvent) GetEventKey() string {
	return d.EventKey
}

func (d *DefaultBaseEvent) EventID() int64 {
	return d.ID
}

func (d *DefaultBaseEvent) EventTopic() string {
	return d.Topic
}

func (d *DefaultBaseEvent) EventTimestamp() int64 {
	return d.Timestamp
}

func (d *DefaultBaseEvent) Payload() interface{} {
	var data interface{}
	if err := json.Unmarshal([]byte(d.Data), &data); err != nil {
		// 处理反序列化错误
		return nil
	}
	return data
}

type EventStore interface {
	SaveEvent(ctx context.Context, event contract.Event) error
	GetEventByID(ctx context.Context, eventID int64) (contract.Event, error)
	GetEvents(ctx context.Context, eventType string, fromID int64, fromTimestamp int64) ([]contract.Event, error)
}

type GormEventStore struct {
	db *gorm.DB
}

func NewGormEventStore(db *gorm.DB) *GormEventStore {
	return &GormEventStore{db: db}
}

func (s *GormEventStore) SaveEvent(ctx context.Context, event contract.Event) error {
	baseEvent := &DefaultBaseEvent{
		Topic:     event.EventTopic(),
		Timestamp: event.EventTimestamp(),
	}
	// 将 Payload 序列化为 JSON
	dataBytes, err := json.Marshal(event.Payload())
	if err != nil {
		return err
	}
	baseEvent.Data = string(dataBytes)
	return s.db.WithContext(ctx).Create(baseEvent).Error
}

func (s *GormEventStore) GetEventByID(ctx context.Context, eventID int64) (contract.Event, error) {
	var event DefaultBaseEvent
	err := s.db.WithContext(ctx).First(&event, eventID).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (s *GormEventStore) GetEvents(ctx context.Context, eventType string, fromID int64, fromTimestamp int64) ([]contract.Event, error) {
	var events []DefaultBaseEvent
	query := s.db.WithContext(ctx).Model(&DefaultBaseEvent{})
	if eventType != "" {
		query = query.Where("type = ?", eventType).Order("id asc")
	}
	if fromID > 0 {
		query = query.Where("id >= ?", fromID).Order("id asc")
	}
	if fromTimestamp > 0 {
		query = query.Where("timestamp >= ?", fromTimestamp).Order("id asc")
	}
	err := query.Find(&events).Error
	if err != nil {
		return nil, err
	}
	// 转换为 contract.Event 接口类型
	result := make([]contract.Event, len(events))
	for i := range events {
		result[i] = &events[i]
	}
	return result, nil
}
