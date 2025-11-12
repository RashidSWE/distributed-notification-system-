package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/zjoart/distributed-notification-system/push-service/internal/models"
	"github.com/zjoart/distributed-notification-system/push-service/pkg/logger"
)

// RabbitMQ wraps RabbitMQ connection and operations
type RabbitMQ struct {
	conn           *amqp091.Connection
	channel        *amqp091.Channel
	url            string
	exchange       string
	pushQueue      string
	failedQueue    string
	prefetchCount  int
	reconnectMutex sync.Mutex
	isConnected    bool
	// messageHandlers []MessageHandler
}

type MessageHandler func(ctx context.Context, msg *models.NotificationMessage) error

func NewRabbitMQ(url, exchange, pushQueue, failedQueue string, prefetchCount int) (*RabbitMQ, error) {
	logger.Info("initializing rabbitmq connection")

	rmq := &RabbitMQ{
		url:           url,
		exchange:      exchange,
		pushQueue:     pushQueue,
		failedQueue:   failedQueue,
		prefetchCount: prefetchCount,
		isConnected:   false,
	}

	if err := rmq.connect(); err != nil {
		return nil, err
	}

	return rmq, nil
}

func (r *RabbitMQ) connect() error {
	r.reconnectMutex.Lock()
	defer r.reconnectMutex.Unlock()

	var err error

	r.conn, err = amqp091.Dial(r.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	r.channel, err = r.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	if err := r.channel.Qos(r.prefetchCount, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	if err := r.channel.ExchangeDeclare(
		r.exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	if _, err := r.channel.QueueDeclare(
		r.pushQueue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare push queue: %w", err)
	}

	if err := r.channel.QueueBind(
		r.pushQueue,
		r.pushQueue,
		r.exchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind push queue: %w", err)
	}

	if _, err := r.channel.QueueDeclare(
		r.failedQueue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare failed queue: %w", err)
	}

	if err := r.channel.QueueBind(
		r.failedQueue,
		r.failedQueue,
		r.exchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind failed queue: %w", err)
	}

	r.isConnected = true

	logger.Info("Connected to RabbitMQ successfully", logger.Fields{
		"exchange":     r.exchange,
		"push_queue":   r.pushQueue,
		"failed_queue": r.failedQueue,
	})

	go r.handleReconnection()

	return nil
}

func (r *RabbitMQ) handleReconnection() {
	connClose := r.conn.NotifyClose(make(chan *amqp091.Error))

	for closeErr := range connClose {
		if closeErr != nil {
			r.isConnected = false
			logger.Error("RabbitMQ connection closed", logger.Fields{
				"error": closeErr.Error(),
			})

			// attempt to reconnect
			for {
				logger.Info("Attempting to reconnect to RabbitMQ...")
				if err := r.connect(); err != nil {
					logger.Error("Failed to reconnect to RabbitMQ", logger.Fields{
						"error": err.Error(),
					})

					time.Sleep(5 * time.Second)
					continue
				}
				logger.Info("Reconnected to RabbitMQ successfully")
				break
			}
		}
	}
}

func (r *RabbitMQ) Consume(ctx context.Context, handler MessageHandler) error {
	msgs, err := r.channel.Consume(
		r.pushQueue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	logger.Info("Started consuming messages from push queue")

	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Info("Stopping message consumption")
				return
			case msg, ok := <-msgs:
				if !ok {
					logger.Warn("Message channel closed, attempting to reconnect")
					return
				}

				var notification models.NotificationMessage
				if err := json.Unmarshal(msg.Body, &notification); err != nil {
					logger.Error("Failed to unmarshal message", logger.Merge(logger.WithError(err), logger.Fields{"message": msg.Body}))

					// reject and don't requeue invalid messages
					msg.Nack(false, false)
					continue
				}

				logDetails := logger.Merge(
					logger.WithUserID(notification.UserID),
					logger.WithNotificationID(notification.ID),
				)

				logger.Info("Received notification message", logDetails)

				// process message
				if err := handler(ctx, &notification); err != nil {
					logger.Error("Failed to process message",
						logger.Merge(logDetails, logger.WithError(err)),
					)

					// handle rate limit errors and send to failed queue without retry
					if errors.Is(err, models.ErrRateLimitExceeded) {
						r.PublishFailed(ctx, &notification, err.Error())
						msg.Ack(false)
						logger.Warn("Rate limited message sent to failed queue", logDetails)
						continue
					}

					// sends the message to failed queue, retry logic has been handled in the service layer
					r.PublishFailed(ctx, &notification, err.Error())
					msg.Ack(false)

					logger.Warn("Message failed after retries, sent to failed queue", logDetails)
				} else {

					msg.Ack(false)
					logger.Info("Message processed successfully", logger.Fields{
						"notification_id": notification.ID,
					})
				}
			}
		}
	}()

	return nil
}

func (r *RabbitMQ) Publish(ctx context.Context, queueName string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = r.channel.PublishWithContext(
		ctx,
		r.exchange,
		queueName, // routing key
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (r *RabbitMQ) PublishFailed(ctx context.Context, notification *models.NotificationMessage, reason string) error {
	failedMsg := models.FailedMessage{
		OriginalMessage: *notification,
		Reason:          reason,
		FailedAt:        time.Now(),
		LastError:       reason,
	}

	logDetails := logger.Merge(
		logger.Fields{
			"reason": reason,
		},
		logger.WithNotificationID(notification.ID),
	)

	logger.Warn("Publishing message to dead letter queue", logDetails)

	return r.Publish(ctx, r.failedQueue, failedMsg)
}

func (r *RabbitMQ) Health() error {
	if !r.isConnected || r.conn == nil || r.conn.IsClosed() {
		return fmt.Errorf("RabbitMQ connection is closed")
	}
	return nil
}

func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			return err
		}
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			return err
		}
	}
	r.isConnected = false
	return nil
}
