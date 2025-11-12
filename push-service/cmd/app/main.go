package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zjoart/distributed-notification-system/push-service/internal/cache"
	"github.com/zjoart/distributed-notification-system/push-service/internal/config"
	handler "github.com/zjoart/distributed-notification-system/push-service/internal/handlers"
	"github.com/zjoart/distributed-notification-system/push-service/internal/push"
	"github.com/zjoart/distributed-notification-system/push-service/internal/queue"
	"github.com/zjoart/distributed-notification-system/push-service/internal/server"
	"github.com/zjoart/distributed-notification-system/push-service/internal/service"
	"github.com/zjoart/distributed-notification-system/push-service/pkg/logger"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found", logger.WithError(err))
	}

	cfg := config.Load()

	redisCache, err := cache.NewRedisCache(cfg.GetRedisAddr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", logger.WithError(err))
	}

	defer redisCache.Close()
	logger.Info("Redis connected successfully")

	rabbitMQ, err := queue.NewRabbitMQ(
		cfg.GetRabbitMQURL(),
		cfg.RabbitMQ.Exchange,
		cfg.RabbitMQ.PushQueue,
		cfg.RabbitMQ.FailedQueue,
		cfg.RabbitMQ.StatusQueue,
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
		logger.Fatal("Failed to initialize FCM service", logger.WithError(err))
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
		cfg.RateLimit,
		rabbitMQ,
	)

	healthHandler := handler.NewHealthHandler(rabbitMQ, redisCache)
	notificationHandler := handler.NewNotificationHandler(notificationService, rabbitMQ)

	httpServer := server.NewServer(
		cfg.Server.Host,
		cfg.Server.Port,
		healthHandler,
		notificationHandler,
	)

	// start HTTP server in goroutine
	go func() {
		if err := httpServer.Start(); err != nil {
			logger.Fatal("Failed to start HTTP server", logger.WithError(err))
		}
	}()

	consumerCtx, cancelConsumer := context.WithCancel(context.Background())
	if err := rabbitMQ.Consume(consumerCtx, notificationService.ProcessNotification); err != nil {
		logger.Fatal("Failed to start consuming messages", logger.WithError(err))
	}

	logger.Info("Push Service started successfully", logger.Fields{
		"http_port": cfg.Server.Port,
		"queue":     cfg.RabbitMQ.PushQueue,
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Push Service")

	cancelConsumer()

	// shutdown gracefully
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error shutting down HTTP server", logger.WithError(err))
	}

	logger.Info("Push Service stopped")

}
