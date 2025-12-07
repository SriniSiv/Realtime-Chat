package config

import (
	"log"
)

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers       []string
	Topic         string
	ConsumerGroup string
	Enabled       bool
}

// GetKafkaConfig returns Kafka configuration from environment
func GetKafkaConfig() *KafkaConfig {
	// Ensure environment is initialized
	if EnvironmentData == nil {
		InitializeEnv()
	}

	if !EnvironmentData.KafkaEnabled {
		log.Println("Kafka is disabled (set KAFKA_ENABLED=true to enable)")
		return &KafkaConfig{Enabled: false}
	}

	config := &KafkaConfig{
		Brokers:       EnvironmentData.KafkaBrokers,
		Topic:         EnvironmentData.KafkaTopic,
		ConsumerGroup: EnvironmentData.KafkaConsumerGroup,
		Enabled:       true,
	}

	log.Printf("Kafka config: brokers=%v, topic=%s, consumerGroup=%s",
		config.Brokers, config.Topic, config.ConsumerGroup)

	return config
}
