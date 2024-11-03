package contract

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/Superdanda/hade/framework"
	"strings"
)

const KafkaKey = "hade:kafka"

// KafkaOption 代表初始化时的选项
type KafkaOption func(container framework.Container, config *KafkaConfig) error

// KafkaService 表示一个 Kafka 服务
type KafkaService interface {
	// GetProducer 获取 Kafka 同步生产者实例
	GetProducer(option ...KafkaOption) (sarama.SyncProducer, error)
	// GetAsyncProducer 获取 Kafka 异步生产者实例
	GetAsyncProducer(option ...KafkaOption) (sarama.AsyncProducer, error)
	// GetConsumer 获取 Kafka 消费者实例
	GetConsumer(option ...KafkaOption) (sarama.Consumer, error)
	// GetConsumerGroup 获取 Kafka 消费者组实例
	GetConsumerGroup(groupID string, topics []string, option ...KafkaOption) (sarama.ConsumerGroup, error)
}

// KafkaConfig 为 Kafka 定义的配置结构
type KafkaConfig struct {
	// 基础配置
	Brokers []string // Kafka broker 列表
	//// 生产者配置
	//Producer *sarama.Config // 生产者配置
	//// 消费者配置
	//Consumer *sarama.Config // 消费者配置
	//// 消费者组配置
	//ConsumerGroup *sarama.Config // 消费者组配置

	ClientConfig *sarama.Config // kafka 配置
}

// UniqKey 用来唯一标识一个 KafkaConfig 配置
func (config *KafkaConfig) UniqKey() string {
	return fmt.Sprintf("%s_%s", strings.Join(config.Brokers, ","))
}
