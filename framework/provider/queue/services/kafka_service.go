package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
	"github.com/Superdanda/hade/framework/provider/queue/queue_base"
	"github.com/google/uuid"
	"log"
	"time"
)

type KafkaQueueService struct {
	container                     framework.Container
	kafkaService                  contract.KafkaService
	RegisterSubscribed            map[string][]contract.EventHandler
	RegisterSubscribedWithContext map[string][]contract.EventHandlerWithContext
	context                       context.Context
	log                           log.Logger
}

// GetContext 为订阅设置上下文
func (k *KafkaQueueService) GetContext() context.Context {
	return k.context
}

func (k *KafkaQueueService) SetContext(ctx context.Context) {
	k.context = ctx
}

func (k *KafkaQueueService) RegisterSubscribe(topic string, handler func(event contract.Event) error) error {
	k.RegisterSubscribed[topic] = append(k.RegisterSubscribed[topic], handler)
	return nil
}

func (k *KafkaQueueService) RegisterSubscribeWithContext(topic string, handler contract.EventHandlerWithContext) error {
	k.RegisterSubscribedWithContext[topic] = append(k.RegisterSubscribedWithContext[topic], handler)
	return nil
}

func (k *KafkaQueueService) GetRegisterSubscribe(topic string) []contract.EventHandler {
	return k.RegisterSubscribed[topic]
}

func NewKafkaQueueService(params ...interface{}) (interface{}, error) {
	kafkaQueueService := &KafkaQueueService{}
	kafkaQueueService.container = params[0].(framework.Container)
	kafkaService := kafkaQueueService.container.MustMake(contract.KafkaKey).(contract.KafkaService)
	kafkaQueueService.kafkaService = kafkaService
	kafkaQueueService.RegisterSubscribed = make(map[string][]contract.EventHandler)
	return kafkaQueueService, nil
}

type KafkaEvent struct {
	EventKey    string                  `json:"eventKey"`  // 事件唯一标识
	Topic       string                  `json:"topic"`     // 事件主题
	Timestamp   int64                   `json:"timestamp"` // 事件时间戳
	Source      string                  `json:"source"`
	Payload     interface{}             `json:"payload"` // 事件负载
	AuthSession *queue_base.AuthSession `json:"authSession"`
}

func NewKafkaEvent(topic, source string, payload interface{}) *KafkaEvent {
	return &KafkaEvent{
		EventKey:  uuid.New().String(), // 使用 UUID 作为事件 ID
		Topic:     topic,
		Timestamp: time.Now().Unix(),
		Source:    source,
		Payload:   payload,
	}
}

func NewKafkaEventByMsg(msgValue string) *KafkaEvent {
	// 定义一个 KafkaEvent 对象
	kafkaEvent := &KafkaEvent{}
	// 使用 json.Unmarshal 将 JSON 字符串解析为 KafkaEvent 对象
	err := json.Unmarshal([]byte(msgValue), kafkaEvent)
	if err != nil {
		return nil
	}
	// 返回 KafkaEvent 对象
	return kafkaEvent
}

func convertToMessage(e contract.Event) (*sarama.ProducerMessage, error) {
	_, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	// 创建Kafka消息
	data, err := json.Marshal(e)
	message := &sarama.ProducerMessage{
		Topic: e.EventTopic(),
		Key:   sarama.StringEncoder(e.GetEventKey()),
		Value: sarama.StringEncoder(data), // 假设payload是字符串
	}
	return message, nil
}

func (k *KafkaQueueService) ProcessSubscribe() {
	for topic, handlers := range k.RegisterSubscribed {
		for _, handler := range handlers {
			k.SubscribeEvent(k.context, topic, handler)
		}
	}

	for topic, contextHandlers := range k.RegisterSubscribedWithContext {
		for _, contextHandler := range contextHandlers {
			k.SubscribeEventWithContext(k.context, topic, contextHandler)
		}
	}
}

// GetEventKey 实现 EventID 方法
func (e *KafkaEvent) GetEventKey() string {
	return e.EventKey
}

// EventTopic 实现 EventTopic 方法
func (e *KafkaEvent) EventTopic() string {
	return e.Topic
}

// EventTimestamp 实现 EventTimestamp 方法
func (e *KafkaEvent) EventTimestamp() int64 {
	return e.Timestamp
}

// EventPayload 实现 EventPayload 方法
func (e *KafkaEvent) EventPayload() interface{} {
	return e.Payload
}

// EventSource 实现 EventSource 方法
func (e *KafkaEvent) EventSource() string {
	return e.Source
}

func (e *KafkaEvent) Create() queue_base.AuthIdentity {
	return e.AuthSession
}

func (e *KafkaEvent) MarshalBinary() ([]byte, error) {
	return json.Marshal(e)
}

func (e *KafkaEvent) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, e)
}

func (k KafkaQueueService) PublishEvent(ctx context.Context, event contract.Event) error {
	producer, err := k.kafkaService.GetProducer()
	if err != nil {
		return err
	}
	producerMessage, err := convertToMessage(event)
	if err != nil {
		return err
	}
	// 发送消息
	_, _, err = producer.SendMessage(producerMessage)
	return err
}

func (k KafkaQueueService) SubscribeEvent(ctx context.Context, topic string, handler func(event contract.Event) error) error {
	consumer, err := k.kafkaService.GetConsumer()
	if err != nil {
		return err
	}
	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case msg := <-partitionConsumer.Messages():
				event := NewKafkaEventByMsg(string(msg.Value))
				handler(event)
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (k KafkaQueueService) SubscribeEventWithContext(ctx context.Context, topic string, handler contract.EventHandlerWithContext) error {
	consumer, err := k.kafkaService.GetConsumer()
	if err != nil {
		return err
	}
	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case msg := <-partitionConsumer.Messages():
				event := NewKafkaEventByMsg(string(msg.Value))
				// ✅ 创建上下文（包含身份 + 容器 + 父上下文）
				mqCtx := queue_base.NewContext(ctx, k.container, event.Create())

				// ✅ 可选设置链路字段
				mqCtx.WithValue("topic", topic)
				mqCtx.WithValue("kafka.offset", msg.Offset)

				// ✅ 执行处理逻辑
				if err := handler(mqCtx, event); err != nil {
					// TODO：日志记录、错误告警
					k.log.Print(err.Error())
				}
			case <-ctx.Done():
				k.log.Print("Kafka 消费协程退出：接收到 Done 信号")
				return
			}
		}
	}()
	return nil
}

func (k KafkaQueueService) ReplayEvents(ctx context.Context, topic string, fromID string, fromTimestamp int64, handler func(event contract.Event) error) error {
	consumer, err := k.kafkaService.GetConsumer()
	if err != nil {
		return err
	}

	// 使用时间戳来设定偏移量，确保从某个时间点开始消费
	offset := sarama.OffsetOldest // 默认从最早的消息开始消费
	if fromTimestamp > 0 {
		// 设置为从某个时间戳的消息开始
		offset = sarama.OffsetNewest // 这里假设你通过时间戳来决定是否是最新的
	}

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, offset)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case msg := <-partitionConsumer.Messages():
				event := NewKafkaEventByMsg(string(msg.Value))
				handler(event)
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (k KafkaQueueService) GetEventById(ctx context.Context, topic string, eventID string) (contract.Event, error) {
	consumer, err := k.kafkaService.GetConsumer()
	if err != nil {
		return nil, err
	}

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		return nil, err
	}

	var result *KafkaEvent
	go func() {
		for msg := range partitionConsumer.Messages() {
			if string(msg.Key) == eventID {
				result = NewKafkaEventByMsg(string(msg.Value))
				break
			}
		}
	}()

	// 等待事件处理完成
	time.Sleep(2 * time.Second) // 可以改为一个更合理的超时处理逻辑

	if result == nil {
		return nil, fmt.Errorf("event with ID %s not found", eventID)
	}

	return result, nil
}

func (k KafkaQueueService) GetEventByTime(ctx context.Context, topic string, fromTimestamp int64) (contract.Event, error) {
	consumer, err := k.kafkaService.GetConsumer()
	if err != nil {
		return nil, err
	}

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		return nil, err
	}

	var result *KafkaEvent
	go func() {
		for msg := range partitionConsumer.Messages() {
			// 假设消息的时间戳存在Data字段中，进行时间戳筛选
			if msg.Timestamp.Unix() >= fromTimestamp {
				result = NewKafkaEventByMsg(string(msg.Value))
				break
			}
		}
	}()

	// 等待事件处理完成
	time.Sleep(2 * time.Second) // 可以改为一个更合理的超时处理逻辑

	if result == nil {
		return nil, fmt.Errorf("event not found from timestamp %d", fromTimestamp)
	}

	return result, nil
}

func (k KafkaQueueService) Close() error {
	// 如果有关闭消费者的逻辑，或者其他资源清理操作，可以在这里实现
	// 示例：关闭所有消费者
	//for _, consumer := range k.kafkaService.GetConsumers() {
	//	consumer.Close() // 关闭消费者连接
	//}
	return nil
}

func (k KafkaQueueService) NewEventAndPublish(ctx context.Context, topic string, payload interface{}) error {
	// 生成新的事件
	event := NewKafkaEvent(topic, "框架服务", payload)
	return k.PublishEvent(ctx, event)
}
