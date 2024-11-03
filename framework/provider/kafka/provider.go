package kafka

import (
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
)

type KafkaProvider struct{}

func (k KafkaProvider) Register(container framework.Container) framework.NewInstance {
	return NewHadeKafka
}

func (k KafkaProvider) Boot(container framework.Container) error {
	return nil
}

func (k KafkaProvider) IsDefer() bool {
	return false
}

func (k KafkaProvider) Params(container framework.Container) []interface{} {
	return []interface{}{container}
}

func (k KafkaProvider) Name() string {
	return contract.KafkaKey
}
