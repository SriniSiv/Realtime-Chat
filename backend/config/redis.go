package config

import (
	"log"
)

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Channel  string
	Enabled  bool
}

// GetRedisConfig returns Redis configuration from environment
func GetRedisConfig() *RedisConfig {
	// Ensure environment is initialized
	if EnvironmentData == nil {
		InitializeEnv()
	}

	if !EnvironmentData.RedisEnabled {
		log.Println("Redis Pub/Sub is disabled (set REDIS_ENABLED=true to enable)")
		return &RedisConfig{Enabled: false}
	}

	config := &RedisConfig{
		Addr:     EnvironmentData.RedisAddr,
		Password: EnvironmentData.RedisPassword,
		DB:       EnvironmentData.RedisDB,
		Channel:  EnvironmentData.RedisChannel,
		Enabled:  true,
	}

	log.Printf("Redis config: addr=%s, channel=%s", config.Addr, config.Channel)

	return config
}
