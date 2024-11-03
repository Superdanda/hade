package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
	"sync"
)

type HadeKafka struct {
	container      framework.Container             // 服务容器
	clients        map[string]sarama.Client        // key为uniqKey, value为redis.Client (连接池）
	syncProducers  map[string]sarama.SyncProducer  // key为uniqKey, value为redis.SyncProducer (同步生产者）
	asyncProducers map[string]sarama.AsyncProducer // key为uniqKey, value为redis.AsyncProducer (异步生产者）
	consumers      map[string]sarama.Consumer      // key为uniqKey, value为redis.Consumer (消费者）
	consumerGroups map[string]sarama.ConsumerGroup // key为uniqKey, value为redis.ConsumerGroup (消费者组）
	lock           *sync.RWMutex
}

func NewHadeKafka(params ...interface{}) (interface{}, error) {
	hadeKafka := &HadeKafka{}
	container := params[0]
	hadeKafka.container = container.(framework.Container)
	hadeKafka.clients = make(map[string]sarama.Client)
	hadeKafka.lock = new(sync.RWMutex)
	return hadeKafka, nil
}

func (k HadeKafka) GetClient(option ...contract.KafkaOption) (sarama.Client, error) {
	// 读取默认配置
	config := GetBaseConfig(k.container)
	// option对opt进行修改
	for _, opt := range option {
		if err := opt(k.container, config); err != nil {
			return nil, err
		}
	}
	uniqKey := config.UniqKey()
	// 判断是否已经实例化了kafka.Client
	k.lock.RLock()
	if client, ok := k.clients[uniqKey]; ok {
		k.lock.RUnlock()
		return client, nil
	}
	k.lock.RUnlock()
	// 没有实例化kafka.Client，那么就要进行实例化操作
	k.lock.Lock()
	defer k.lock.Unlock()
	client, err := sarama.NewClient(config.Brokers, config.ClientConfig)
	if err != nil {
		return nil, err
	}
	k.clients[uniqKey] = client
	return client, err
}

func (k HadeKafka) GetProducer(option ...contract.KafkaOption) (sarama.SyncProducer, error) {

	// 读取默认配置
	config := GetBaseConfig(k.container)
	// option对opt进行修改
	for _, opt := range option {
		if err := opt(k.container, config); err != nil {
			return nil, err
		}
	}
	uniqKey := config.UniqKey()

	// 判断是否已经实例化了kafka.Client
	k.lock.RLock()
	if producer, ok := k.syncProducers[uniqKey]; ok {
		k.lock.Lock()
		return producer, nil
	}

	k.lock.RUnlock()

	client, err := k.GetClient(option...)
	if err != nil {
		return nil, err
	}

	k.lock.Lock()
	defer k.lock.Unlock()
	syncProducer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}
	k.syncProducers[uniqKey] = syncProducer
	return syncProducer, nil
}

func (k HadeKafka) GetAsyncProducer(option ...contract.KafkaOption) (sarama.AsyncProducer, error) {
	// 读取默认配置
	config := GetBaseConfig(k.container)
	// 应用 options 配置修改
	for _, opt := range option {
		if err := opt(k.container, config); err != nil {
			return nil, err
		}
	}
	uniqKey := config.UniqKey()

	// 检查是否已实例化 asyncProducer
	k.lock.RLock()
	if producer, ok := k.asyncProducers[uniqKey]; ok {
		k.lock.RUnlock()
		return producer, nil
	}
	k.lock.RUnlock()

	// 获取 client 实例
	client, err := k.GetClient(option...)
	if err != nil {
		return nil, err
	}

	// 创建并保存 asyncProducer
	k.lock.Lock()
	defer k.lock.Unlock()
	asyncProducer, err := sarama.NewAsyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}
	k.asyncProducers[uniqKey] = asyncProducer
	return asyncProducer, nil
}

func (k HadeKafka) GetConsumer(option ...contract.KafkaOption) (sarama.Consumer, error) {
	// 读取默认配置
	config := GetBaseConfig(k.container)
	// 应用 options 配置修改
	for _, opt := range option {
		if err := opt(k.container, config); err != nil {
			return nil, err
		}
	}
	uniqKey := config.UniqKey()

	// 检查是否已实例化 consumer
	k.lock.RLock()
	if consumer, ok := k.consumers[uniqKey]; ok {
		k.lock.RUnlock()
		return consumer, nil
	}
	k.lock.RUnlock()

	// 获取 client 实例
	client, err := k.GetClient(option...)
	if err != nil {
		return nil, err
	}

	// 创建并保存 consumer
	k.lock.Lock()
	defer k.lock.Unlock()
	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		return nil, err
	}
	k.consumers[uniqKey] = consumer
	return consumer, nil
}

func (k HadeKafka) GetConsumerGroup(groupID string, topics []string, option ...contract.KafkaOption) (sarama.ConsumerGroup, error) {
	// 读取默认配置
	config := GetBaseConfig(k.container)
	// 应用 options 配置修改
	for _, opt := range option {
		if err := opt(k.container, config); err != nil {
			return nil, err
		}
	}
	uniqKey := config.UniqKey() + "_" + groupID

	// 检查是否已实例化 consumerGroup
	k.lock.RLock()
	if group, ok := k.consumerGroups[uniqKey]; ok {
		k.lock.RUnlock()
		return group, nil
	}
	k.lock.RUnlock()

	// 获取 client 实例
	client, err := k.GetClient(option...)
	if err != nil {
		return nil, err
	}

	// 创建并保存 consumerGroup
	k.lock.Lock()
	defer k.lock.Unlock()
	consumerGroup, err := sarama.NewConsumerGroupFromClient(groupID, client)
	if err != nil {
		return nil, err
	}
	k.consumerGroups[uniqKey] = consumerGroup
	return consumerGroup, nil
}
