package config

import (
	"fmt"
	"os"
	"strconv"
)

// app configuration
type Config struct {
	Server           ServerConfig
	RabbitMQ         RabbitMQConfig
	Redis            RedisConfig
	FCM              FCMConfig
	Circuit          CircuitBreakerConfig
	Retry            RetryConfig
	RateLimit        RateLimitConfig
	ExternalServices ExternalServicesConfig
}

// server configuration
type ServerConfig struct {
	Host string
	Port int
}

// rabbitMQ connection settings
type RabbitMQConfig struct {
	Host          string
	Port          int
	User          string
	Password      string
	Exchange      string
	PushQueue     string
	FailedQueue   string
	StatusQueue   string
	PrefetchCount int
}

// redis connection settings
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// Firebase Cloud Messaging configuration
type FCMConfig struct {
	ProjectID       string
	CredentialsPath string
	Timeout         int // seconds
}

// circuit breaker settings
type CircuitBreakerConfig struct {
	MaxRequests      uint32
	Interval         int // seconds
	Timeout          int // seconds
	FailureThreshold uint32
}

// retry policy configuration
type RetryConfig struct {
	MaxAttempts     int
	InitialInterval int // seconds
	MaxInterval     int // seconds
	Multiplier      float64
}

// rate limiting configuration
type RateLimitConfig struct {
	Requests int // max requests per window
	Window   int // window duration in seconds
}

// external services configuration
type ExternalServicesConfig struct {
	TemplateServiceURL string
}

func Load() *Config {
	config := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST"),
			Port: getEnvAsInt("SERVER_PORT"),
		},
		RabbitMQ: RabbitMQConfig{
			Host:          getEnv("RABBITMQ_HOST"),
			Port:          getEnvAsInt("RABBITMQ_PORT"),
			User:          getEnv("RABBITMQ_USER"),
			Password:      getEnv("RABBITMQ_PASS"),
			Exchange:      getEnv("RABBITMQ_EXCHANGE"),
			PushQueue:     getEnv("RABBITMQ_PUSH_QUEUE"),
			FailedQueue:   getEnv("RABBITMQ_FAILED_QUEUE"),
			StatusQueue:   getEnv("RABBITMQ_STATUS_QUEUE"),
			PrefetchCount: getEnvAsInt("RABBITMQ_PREFETCH_COUNT"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST"),
			Port:     getEnvAsInt("REDIS_PORT"),
			Password: "", // getEnv("REDIS_PASSWORD"),
			DB:       getEnvAsInt("REDIS_DB"),
		},
		FCM: FCMConfig{
			ProjectID:       getEnv("FCM_PROJECT_ID"),
			CredentialsPath: getEnv("FCM_CREDENTIALS_FILE"),
			Timeout:         getEnvAsInt("FCM_TIMEOUT"),
		},
		Circuit: CircuitBreakerConfig{
			MaxRequests:      uint32(getEnvAsInt("CIRCUIT_MAX_REQUESTS")),
			Interval:         getEnvAsInt("CIRCUIT_INTERVAL"),
			Timeout:          getEnvAsInt("CIRCUIT_TIMEOUT"),
			FailureThreshold: uint32(getEnvAsInt("CIRCUIT_FAILURE_THRESHOLD")),
		},
		Retry: RetryConfig{
			MaxAttempts:     getEnvAsInt("RETRY_MAX_ATTEMPTS"),
			InitialInterval: getEnvAsInt("RETRY_INITIAL_INTERVAL"),
			MaxInterval:     getEnvAsInt("RETRY_MAX_INTERVAL"),
			Multiplier:      getEnvAsFloat("RETRY_MULTIPLIER"),
		},
		RateLimit: RateLimitConfig{
			Requests: getEnvAsInt("RATE_LIMIT_REQUESTS"),
			Window:   getEnvAsInt("RATE_LIMIT_WINDOW"),
		},
		ExternalServices: ExternalServicesConfig{
			TemplateServiceURL: getEnv("TEMPLATE_SERVICE_URL"),
		},
	}

	return config
}

func (c *Config) GetRabbitMQURL() string {
	return fmt.Sprintf(
		"amqp://%s:%s@%s:%d/",
		c.RabbitMQ.User,
		c.RabbitMQ.Password,
		c.RabbitMQ.Host,
		c.RabbitMQ.Port,
	)
}

func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

func getEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	panic(fmt.Sprintf("%s is required", key))
}

func getEnvAsInt(key string) int {
	valueStr := os.Getenv(key)

	value, err := strconv.Atoi(valueStr)

	if err != nil {
		panic(fmt.Sprintf("Int key error: %s", err.Error()))
	}

	return value
}

func getEnvAsFloat(key string) float64 {
	valueStr := os.Getenv(key)

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		panic(fmt.Sprintf("Float key error: %s", err.Error()))
	}

	return value
}
