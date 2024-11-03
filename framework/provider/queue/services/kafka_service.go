package services

import (
	"context"
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
)

type KafkaQueueService struct {
	container    framework.Container
	kafkaService contract.KafkaService
}

func NewKafkaQueueService(params ...interface{}) (interface{}, error) {
	kafkaQueueService := &KafkaQueueService{}
	kafkaQueueService.container = params[0].(framework.Container)
	kafkaService := kafkaQueueService.container.MustMake(contract.KafkaKey).(contract.KafkaService)
	kafkaQueueService.kafkaService = kafkaService
	return kafkaQueueService, nil
}

func (k KafkaQueueService) PublishEvent(ctx context.Context, event contract.Event) error {
	//TODO implement me
	panic("implement me")
}

func (k KafkaQueueService) SubscribeEvent(ctx context.Context, topic string, handler func(event contract.Event) error) error {
	//TODO implement me
	panic("implement me")
}

func (k KafkaQueueService) ReplayEvents(ctx context.Context, topic string, fromID string, fromTimestamp int64, handler func(event contract.Event) error) error {
	//TODO implement me
	panic("implement me")
}

func (k KafkaQueueService) GetEventById(ctx context.Context, topic string, eventID string) (contract.Event, error) {
	//TODO implement me
	panic("implement me")
}

func (k KafkaQueueService) GetEventByTime(ctx context.Context, topic string, fromTimestamp int64) (contract.Event, error) {
	//TODO implement me
	panic("implement me")
}

func (k KafkaQueueService) Close() error {
	//TODO implement me
	panic("implement me")
}

func (k KafkaQueueService) NewEventAndPublish(ctx context.Context, topic string, payload interface{}) error {
	//TODO implement me
	panic("implement me")
}
