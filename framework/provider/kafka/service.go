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

func (k HadeKafka) GetClientDefault() (sarama.Client, error) {
	// 读取默认配置
	config := GetBaseConfig(k.container)
	return k.getClientByConfig(config)
}

func (k HadeKafka) GetConsumerDefault() (sarama.Consumer, error) {
	config := GetBaseConfig(k.container)
	return k.getConsumer(k.GetClientDefault, config)
}

func (k HadeKafka) GetConsumerGroupDefault(groupID string, topics []string) (sarama.ConsumerGroup, error) {
	config := GetBaseConfig(k.container)
	return k.getConsumerGroup(k.GetClientDefault, config, groupID, topics)
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
	return k.getClientByConfig(config)
}

// 内部方法：获取消费者组
func (k HadeKafka) getConsumerGroup(
	getClientFunc func() (sarama.Client, error),
	config *contract.KafkaConfig,
	groupID string,
	topics []string) (sarama.ConsumerGroup, error) {

	uniqKey := config.UniqKey() + "_" + groupID
	k.lock.RLock()
	if group, ok := k.consumerGroups[uniqKey]; ok {
		k.lock.RUnlock()
		return group, nil
	}
	k.lock.RUnlock()
	client, err := getClientFunc()
	if err != nil {
		return nil, err
	}
	k.lock.Lock()
	defer k.lock.Unlock()
	consumerGroup, err := sarama.NewConsumerGroupFromClient(groupID, client)
	if err != nil {
		return nil, err
	}
	k.consumerGroups[uniqKey] = consumerGroup
	return consumerGroup, nil
}

// 内部方法：获取消费者
func (k HadeKafka) getConsumer(
	getClientFunc func() (sarama.Client, error),
	config *contract.KafkaConfig) (sarama.Consumer, error) {

	uniqKey := config.UniqKey()
	k.lock.RLock()
	if consumer, ok := k.consumers[uniqKey]; ok {
		k.lock.RUnlock()
		return consumer, nil
	}
	k.lock.RUnlock()
	client, err := getClientFunc()
	if err != nil {
		return nil, err
	}
	k.lock.Lock()
	defer k.lock.Unlock()
	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		return nil, err
	}
	k.consumers[uniqKey] = consumer
	return consumer, nil
}
func NewHadeKafka(params ...interface{}) (interface{}, error) {
	hadeKafka := &HadeKafka{}
	container := params[0]
	hadeKafka.container = container.(framework.Container)
	hadeKafka.clients = make(map[string]sarama.Client)
	hadeKafka.syncProducers = make(map[string]sarama.SyncProducer)
	hadeKafka.asyncProducers = make(map[string]sarama.AsyncProducer)
	hadeKafka.consumers = make(map[string]sarama.Consumer)
	hadeKafka.consumerGroups = make(map[string]sarama.ConsumerGroup)
	hadeKafka.lock = new(sync.RWMutex)
	return hadeKafka, nil
}

func (k HadeKafka) getClientByConfig(config *contract.KafkaConfig) (sarama.Client, error) {
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
	return k.getProducer(func() (sarama.Client, error) {
		return k.GetClient(option...)
	}, config)
}

func (k HadeKafka) GetProducerDefault() (sarama.SyncProducer, error) {
	// 读取默认配置
	config := GetBaseConfig(k.container)
	return k.getProducer(k.GetClientDefault, config)
}

func (k HadeKafka) GetAsyncProducerDefault() (sarama.AsyncProducer, error) {
	config := GetBaseConfig(k.container)
	return k.getAsyncProducer(k.GetClientDefault, config)
}

// 内部方法：获取异步生产者
func (k HadeKafka) getAsyncProducer(
	getClientFunc func() (sarama.Client, error),
	config *contract.KafkaConfig) (sarama.AsyncProducer, error) {

	uniqKey := config.UniqKey()
	k.lock.RLock()
	if producer, ok := k.asyncProducers[uniqKey]; ok {
		k.lock.RUnlock()
		return producer, nil
	}
	k.lock.RUnlock()
	client, err := getClientFunc()
	if err != nil {
		return nil, err
	}
	k.lock.Lock()
	defer k.lock.Unlock()
	asyncProducer, err := sarama.NewAsyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}
	k.asyncProducers[uniqKey] = asyncProducer
	return asyncProducer, nil
}

func (k HadeKafka) getProducer(
	getClientFunc func() (sarama.Client, error), // 传入一个返回 sarama.Client 的函数
	config *contract.KafkaConfig) (sarama.SyncProducer, error) {
	uniqKey := config.UniqKey()
	// 判断是否已经实例化了kafka.Client
	if producer, ok := k.syncProducers[uniqKey]; ok {
		return producer, nil
	}
	client, err := getClientFunc()
	if err != nil {
		return nil, err
	}
	k.lock.RLock()
	defer k.lock.RUnlock()
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
	return k.getAsyncProducer(func() (sarama.Client, error) {
		return k.GetClient(option...)
	}, config)
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
	return k.getConsumer(func() (sarama.Client, error) {
		return k.GetClient(option...)
	}, config)
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
	return k.getConsumerGroup(
		func() (sarama.Client, error) {
			return k.GetClient(option...)
		}, config, groupID, topics,
	)
}
