package queue

import (
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
	"github.com/Superdanda/hade/framework/provider/queue/services"
)

type QueueProvider struct {
}

func (q QueueProvider) Register(container framework.Container) framework.NewInstance {
	configService := container.MustMake(contract.ConfigKey).(contract.Config)
	queueDriver := configService.GetString("queue.driver")

	if queueDriver == "" {
		queueDriver = "default"
	}

	switch queueDriver {
	case "default":
		return services.NewMemoryQueueService
	case "kafka":
		return services.NewKafkaQueueService
	default:
		return services.NewMemoryQueueService
	}
}

func (q QueueProvider) Boot(container framework.Container) error {
	return nil
}

func (q QueueProvider) IsDefer() bool {
	return false
}

func (q QueueProvider) Params(container framework.Container) []interface{} {
	return []interface{}{container}
}

func (q QueueProvider) Name() string {
	return contract.QueueKey
}
