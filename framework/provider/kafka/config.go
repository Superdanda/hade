package kafka

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/Superdanda/hade/framework"
	"github.com/Superdanda/hade/framework/contract"
	"strconv"
	"time"
)

func GetBaseConfig(c framework.Container) *contract.KafkaConfig {
	logService := c.MustMake(contract.LogKey).(contract.Log)
	config := c.MustMake(contract.ConfigKey).(contract.Config)
	kafkaName := config.GetString("kafka.name")
	kafkaConfig := &contract.KafkaConfig{}
	opt := WithConfigPath(kafkaName)
	err := opt(c, kafkaConfig)
	if err != nil {
		logService.Error(context.Background(), "kafka连接失败，请检查配置是否正确", nil)
		return nil
	}
	return kafkaConfig
}

func WithConfigPath(configPath string) contract.KafkaOption {
	return func(container framework.Container, config *contract.KafkaConfig) error {
		configService := container.MustMake(contract.ConfigKey).(contract.Config)
		baseConfigPath := configPath + ".base"
		producerConfigPath := configPath + ".producer"
		consumerConfigPath := configPath + ".consumer"
		consumerGroupConfigPath := configPath + ".consumerGroup"

		baseConfMap := configService.GetStringMapString(baseConfigPath)
		producerConfMap := configService.GetStringMapString(producerConfigPath)
		consumerConfMap := configService.GetStringMapString(consumerConfigPath)
		consumerGroupConfMap := configService.GetStringMapString(consumerGroupConfigPath)

		brokers := configService.GetStringSlice(baseConfigPath + ".brokers")
		config.Brokers = brokers

		saramaConfig := &sarama.Config{}

		version, ok := baseConfMap["version"]
		if ok {
			kafkaVersion, err := sarama.ParseKafkaVersion(version)
			if err != nil {
				fmt.Println("Error parsing kafka version: ", err)
			}
			saramaConfig.Version = kafkaVersion
		}

		// 配置 Admin (管理接口)
		if retryMax, ok := baseConfMap["admin_retry_max"]; ok {
			if max, err := strconv.Atoi(retryMax); err == nil {
				saramaConfig.Admin.Retry.Max = max
			}
		}

		if retryBackoff, ok := baseConfMap["admin_retry_backoff"]; ok {
			if backoff, err := time.ParseDuration(retryBackoff); err == nil {
				saramaConfig.Admin.Retry.Backoff = backoff
			}
		}
		if timeout, ok := baseConfMap["admin_timeout"]; ok {
			if duration, err := time.ParseDuration(timeout); err == nil {
				saramaConfig.Admin.Timeout = duration
			}
		}

		// 配置 Net (网络)
		if maxOpenRequests, ok := baseConfMap["net_max_open_requests"]; ok {
			if max, err := strconv.Atoi(maxOpenRequests); err == nil {
				saramaConfig.Net.MaxOpenRequests = max
			}
		}
		if dialTimeout, ok := baseConfMap["net_dial_timeout"]; ok {
			if duration, err := time.ParseDuration(dialTimeout); err == nil {
				saramaConfig.Net.DialTimeout = duration
			}
		}
		if readTimeout, ok := baseConfMap["net_read_timeout"]; ok {
			if duration, err := time.ParseDuration(readTimeout); err == nil {
				saramaConfig.Net.ReadTimeout = duration
			}
		}
		if writeTimeout, ok := baseConfMap["net_write_timeout"]; ok {
			if duration, err := time.ParseDuration(writeTimeout); err == nil {
				saramaConfig.Net.WriteTimeout = duration
			}
		}

		// 配置 TLS (加密连接)
		if enableTLS, ok := baseConfMap["net_tls_enable"]; ok {
			saramaConfig.Net.TLS.Enable, _ = strconv.ParseBool(enableTLS)
		}
		if tlsSkipVerify, ok := baseConfMap["net_tls_insecure_skip_verify"]; ok {
			if skip, err := strconv.ParseBool(tlsSkipVerify); err == nil && saramaConfig.Net.TLS.Enable {
				saramaConfig.Net.TLS.Config = &tls.Config{InsecureSkipVerify: skip}
			}
		}

		// 配置 SASL (认证)
		if enableSASL, ok := baseConfMap["net_sasl_enable"]; ok {
			saramaConfig.Net.SASL.Enable, _ = strconv.ParseBool(enableSASL)
		}
		if saramaConfig.Net.SASL.Enable {
			saramaConfig.Net.SASL.User = baseConfMap["net_sasl_user"]
			saramaConfig.Net.SASL.Password = baseConfMap["net_sasl_password"]
			if saslMechanism, ok := baseConfMap["net_sasl_mechanism"]; ok {
				switch saslMechanism {
				case "PLAIN":
					saramaConfig.Net.SASL.Mechanism = sarama.SASLTypePlaintext
				case "SCRAM":
					saramaConfig.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256 // 可根据需要调整
				}
			}
		}

		// 配置 Producer (生产者)
		if requiredAcks, ok := producerConfMap["required_acks"]; ok {
			if acks, err := strconv.Atoi(requiredAcks); err == nil {
				saramaConfig.Producer.RequiredAcks = sarama.RequiredAcks(acks)
			}
		}
		if retryMax, ok := producerConfMap["retry_max"]; ok {
			if max, err := strconv.Atoi(retryMax); err == nil {
				saramaConfig.Producer.Retry.Max = max
			}
		}
		if returnSuccesses, ok := producerConfMap["return_successes"]; ok {
			saramaConfig.Producer.Return.Successes, _ = strconv.ParseBool(returnSuccesses)
		}
		if compression, ok := producerConfMap["compression"]; ok {
			switch compression {
			case "gzip":
				saramaConfig.Producer.Compression = sarama.CompressionGZIP
			case "snappy":
				saramaConfig.Producer.Compression = sarama.CompressionSnappy
			case "lz4":
				saramaConfig.Producer.Compression = sarama.CompressionLZ4
			case "zstd":
				saramaConfig.Producer.Compression = sarama.CompressionZSTD
			}
		}
		if flushFrequency, ok := producerConfMap["flush_frequency"]; ok {
			if duration, err := time.ParseDuration(flushFrequency); err == nil {
				saramaConfig.Producer.Flush.Frequency = duration
			}
		}

		// 配置 Consumer (消费者)
		if offsetsInitial, ok := consumerConfMap["offsets_initial"]; ok {
			if initial, err := strconv.Atoi(offsetsInitial); err == nil {
				saramaConfig.Consumer.Offsets.Initial = int64(initial)
			}
		}
		if returnErrors, ok := consumerConfMap["return_errors"]; ok {
			saramaConfig.Consumer.Return.Errors, _ = strconv.ParseBool(returnErrors)
		}
		if maxWaitTime, ok := consumerConfMap["max_wait_time"]; ok {
			if duration, err := time.ParseDuration(maxWaitTime); err == nil {
				saramaConfig.Consumer.MaxWaitTime = duration
			}
		}

		// 配置 Consumer Group (消费者组)
		if rebalanceStrategy, ok := consumerGroupConfMap["group_rebalance_strategy"]; ok {
			switch rebalanceStrategy {
			case "range":
				saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
			case "roundrobin":
				saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
			case "sticky":
				saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
			}
		}
		if sessionTimeout, ok := consumerGroupConfMap["group_session_timeout"]; ok {
			if duration, err := time.ParseDuration(sessionTimeout); err == nil {
				saramaConfig.Consumer.Group.Session.Timeout = duration
			}
		}
		if heartbeatInterval, ok := consumerGroupConfMap["group_heartbeat_interval"]; ok {
			if duration, err := time.ParseDuration(heartbeatInterval); err == nil {
				saramaConfig.Consumer.Group.Heartbeat.Interval = duration
			}
		}

		// 将配置应用到 KafkaConfig
		config.ClientConfig = saramaConfig

		return nil
	}
}
