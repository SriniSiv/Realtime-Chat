package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// EnvironmentDataType holds all environment configuration
type EnvironmentDataType struct {
	// Database Configuration
	DBHost     string `validate:"required"`
	DBPort     string `validate:"required"`
	DBUser     string `validate:"required"`
	DBPassword string `validate:"required"`
	DBName     string `validate:"required"`
	DBSSLMode  string

	// JWT Configuration
	AccessTokenSecret  string `validate:"required"`
	RefreshTokenSecret string `validate:"required"`
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration

	// Server Configuration
	ServerPort string
	GinMode    string

	// CORS Configuration
	AllowedOrigins []string

	// Kafka Configuration
	KafkaEnabled       bool
	KafkaBrokers       []string
	KafkaTopic         string
	KafkaConsumerGroup string

	// Redis Configuration
	RedisEnabled  bool
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	RedisChannel  string

	// Instance Configuration
	InstanceID string
}

// EnvironmentData holds the loaded environment variables
var EnvironmentData *EnvironmentDataType

// InitializeEnv loads and initializes environment variables
func InitializeEnv() {
	log.Println("InitializeEnv: Loading environment configuration...")

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Parse durations
	accessExpiry := parseDuration(os.Getenv("ACCESS_TOKEN_EXPIRY"), 15*time.Minute)
	refreshExpiry := parseDuration(os.Getenv("REFRESH_TOKEN_EXPIRY"), 168*time.Hour)

	// Parse Redis DB
	redisDB := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err == nil {
			redisDB = db
		}
	}

	// Get instance ID
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		hostname, _ := os.Hostname()
		instanceID = hostname
	}

	EnvironmentData = &EnvironmentDataType{
		// Database Configuration
		DBHost:     getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:     getEnvOrDefault("DB_PORT", "5432"),
		DBUser:     getEnvOrDefault("DB_USER", "srini"),
		DBPassword: getEnvOrDefault("DB_PASSWORD", "srini@4614"),
		DBName:     getEnvOrDefault("DB_NAME", "realtime_chat"),
		DBSSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),

		// JWT Configuration
		AccessTokenSecret:  getEnvOrDefault("ACCESS_TOKEN_SECRET", "your-access-token-secret-key-change-this-in-production"),
		RefreshTokenSecret: getEnvOrDefault("REFRESH_TOKEN_SECRET", "your-refresh-token-secret-key-change-this-in-production"),
		AccessTokenExpiry:  accessExpiry,
		RefreshTokenExpiry: refreshExpiry,

		// Server Configuration
		ServerPort: getEnvOrDefault("SERVER_PORT", "8080"),
		GinMode:    getEnvOrDefault("GIN_MODE", "debug"),

		// CORS Configuration
		AllowedOrigins: strings.Split(getEnvOrDefault("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173"), ","),

		// Kafka Configuration
		KafkaEnabled:       os.Getenv("KAFKA_ENABLED") == "true",
		KafkaBrokers:       strings.Split(getEnvOrDefault("KAFKA_BROKERS", "localhost:9092"), ","),
		KafkaTopic:         getEnvOrDefault("KAFKA_TOPIC", "chat-messages"),
		KafkaConsumerGroup: getEnvOrDefault("KAFKA_CONSUMER_GROUP", "chat-persistence-group"),

		// Redis Configuration
		RedisEnabled:  os.Getenv("REDIS_ENABLED") == "true",
		RedisAddr:     getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       redisDB,
		RedisChannel:  getEnvOrDefault("REDIS_CHANNEL", "chat:messages"),

		// Instance Configuration
		InstanceID: instanceID,
	}

	log.Printf("Environment loaded: DBHost=%s, DBName=%s, ServerPort=%s, KafkaEnabled=%v, RedisEnabled=%v, InstanceID=%s",
		EnvironmentData.DBHost,
		EnvironmentData.DBName,
		EnvironmentData.ServerPort,
		EnvironmentData.KafkaEnabled,
		EnvironmentData.RedisEnabled,
		EnvironmentData.InstanceID,
	)
}

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseDuration parses a duration string or returns default
func parseDuration(s string, defaultVal time.Duration) time.Duration {
	if s == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultVal
	}
	return d
}

// GetDSN returns the database connection string
func (e *EnvironmentDataType) GetDSN() string {
	return "host=" + e.DBHost +
		" user=" + e.DBUser +
		" password=" + e.DBPassword +
		" dbname=" + e.DBName +
		" port=" + e.DBPort +
		" sslmode=" + e.DBSSLMode
}
