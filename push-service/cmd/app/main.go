package main

import (
	"context"
	"time"

	"github.com/zjoart/distributed-notification-system/push-service/internal/cache"
	"github.com/zjoart/distributed-notification-system/push-service/internal/config"
	"github.com/zjoart/distributed-notification-system/push-service/internal/database"
	"github.com/zjoart/distributed-notification-system/push-service/internal/push"
	"github.com/zjoart/distributed-notification-system/push-service/internal/queue"
	"github.com/zjoart/distributed-notification-system/push-service/internal/service"
	"github.com/zjoart/distributed-notification-system/push-service/pkg/logger"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found", logger.WithError(err))
	}

	cfg := config.Load()

	db, err := database.LoadPostgres(cfg.GetPostgresDSN())

	if err != nil {
		logger.Fatal("Failed to connect to database", logger.Fields{
			"error": err.Error(),
		})
	}

	defer db.Close()
	logger.Info("Database connected successfully")

	redisCache, err := cache.NewRedisCache(cfg.GetRedisAddr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", logger.Fields{
			"error": err.Error(),
		})
	}

	defer redisCache.Close()
	logger.Info("Redis connected successfully")

	rabbitMQ, err := queue.NewRabbitMQ(
		cfg.GetRabbitMQURL(),
		cfg.RabbitMQ.Exchange,
		cfg.RabbitMQ.PushQueue,
		cfg.RabbitMQ.FailedQueue,
		cfg.RabbitMQ.PrefetchCount,
	)
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", logger.Fields{
			"error": err.Error(),
		})
	}
	defer rabbitMQ.Close()
	logger.Info("RabbitMQ connected successfully")

	// initialize the circuit breaker
	circuitBreaker := push.NewCircuitBreaker(
		cfg.Circuit.MaxRequests,
		cfg.Circuit.FailureThreshold,
		time.Duration(cfg.Circuit.Interval)*time.Second,
		time.Duration(cfg.Circuit.Timeout)*time.Second,
	)

	ctx := context.Background()
	fcmService, err := push.NewFCMService(
		ctx,
		cfg.FCM.ProjectID,
		cfg.FCM.CredentialsPath,
		cfg.FCM.Timeout,
		circuitBreaker,
	)

	if err != nil {
		logger.Fatal("Failed to initialize FCM service", logger.Fields{
			"error": err.Error(),
		})
	}
	logger.Info("FCM service initialized successfully")

	retryService := service.NewRetryService(
		cfg.Retry.MaxAttempts,
		cfg.Retry.InitialInterval,
		cfg.Retry.MaxInterval,
		cfg.Retry.Multiplier,
	)

	notificationService := service.NewNotificationService(
		fcmService,
		retryService,
		redisCache,
	)

	consumerCtx, cancelConsumer := context.WithCancel(context.Background())
	if err := rabbitMQ.Consume(consumerCtx, notificationService.ProcessNotification); err != nil {
		logger.Fatal("Failed to start consuming messages", logger.Fields{
			"error": err.Error(),
		})
	}

	cancelConsumer()

}
