package service

import (
	"context"
	"fmt"

	"github.com/zjoart/distributed-notification-system/push-service/internal/cache"
	"github.com/zjoart/distributed-notification-system/push-service/internal/models"
	"github.com/zjoart/distributed-notification-system/push-service/internal/push"
	"github.com/zjoart/distributed-notification-system/push-service/pkg/logger"
)

type NotificationService struct {
	fcmService   *push.FCMService
	retryService *RetryService
	cache        *cache.RedisCache
}

func NewNotificationService(
	fcmService *push.FCMService,
	retryService *RetryService,
	cache *cache.RedisCache,
) *NotificationService {
	return &NotificationService{
		fcmService:   fcmService,
		retryService: retryService,
		cache:        cache,
	}
}

// process notification message
func (s *NotificationService) ProcessNotification(ctx context.Context, msg *models.NotificationMessage) error {

	loggerDetails := logger.Merge(
		logger.WithNotificationID(msg.ID),
		logger.WithUserID(msg.UserID),
	)
	if err := s.checkRateLimit(ctx, msg.UserID); err != nil {
		logger.Warn("Rate limit exceeded", loggerDetails)
		return err
	}

	if err := msg.Validate(); err != nil {
		logger.Error("Invalid notification message",
			logger.Merge(loggerDetails, logger.WithError(err)),
		)

		return err
	}

	if err := s.checkIdempotency(ctx, msg.ID); err != nil {
		logger.Warn("Duplicate notification detected", loggerDetails)

		return nil // return nil to acknowledge the message
	}

	logger.Info("Processing notification", logger.Merge(loggerDetails, logger.Fields{
		"device_count":  len(msg.DeviceTokens),
		"attempt_count": msg.AttemptCount,
	}))

	notification, err := s.prepareNotification(msg)
	if err != nil {
		logger.Error("Failed to prepare notification", logger.Merge(loggerDetails,
			logger.WithError(
				err,
			)))

		return err
	}

	results, err := s.sendNotification(ctx, msg, notification)
	if err != nil {
		logger.Error("Failed to send notification", logger.Merge(loggerDetails,
			logger.WithError(
				err,
			)))
		return err
	}

	successCount := 0
	failedCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			failedCount++
		}
	}

	logger.Info("Notification processing completed", logger.Merge(loggerDetails,
		logger.Fields{
			"success_count": successCount,
			"failed_count":  failedCount,
			"total":         len(results),
		}))

	// mark as processed for idempotency
	s.markAsProcessed(ctx, msg.ID)

	return nil
}

// prepare the notification content
func (s *NotificationService) prepareNotification(msg *models.NotificationMessage) (*models.PushNotification, error) {

	return &models.PushNotification{
		Title:    msg.Title,
		Body:     msg.Body,
		ImageURL: msg.ImageURL,
		Link:     msg.Link,
		Data:     msg.Data,
		Priority: msg.Priority,
	}, nil
}

// send the notification to device tokens
func (s *NotificationService) sendNotification(ctx context.Context, msg *models.NotificationMessage, notification *models.PushNotification) ([]*models.NotificationResult, error) {

	// validate device tokens
	validTokens := make([]string, 0, len(msg.DeviceTokens))
	for _, token := range msg.DeviceTokens {
		if token != "" {
			validTokens = append(validTokens, token)
		}
	}

	if len(validTokens) == 0 {
		return nil, models.ErrNoDeviceTokens
	}

	var results []*models.NotificationResult

	err := s.retryService.RetryWithBackoff(ctx, msg.AttemptCount, func() error {
		var err error

		if len(validTokens) == 1 {
			// send notification to single device
			result, err := s.fcmService.SendNotification(ctx, validTokens[0], notification)
			if err != nil {
				return err
			}
			result.CorrelationID = msg.CorrelationID
			results = []*models.NotificationResult{result}
		} else {
			// send notification to multiple devices
			results, err = s.fcmService.SendToMultipleDevices(ctx, validTokens, notification)

			if err != nil {
				return err
			}

			// add correlation ID to results
			for _, result := range results {
				result.CorrelationID = msg.CorrelationID
			}
		}

		// check if all sends failed
		allFailed := true
		for _, result := range results {
			if result.Success {
				allFailed = false
				break
			}
		}

		if allFailed {
			return fmt.Errorf("all notification sends failed")
		}

		return nil
	})

	// Increment attempt count
	msg.IncrementAttempt()

	return results, err
}

// check if notification has already been processed
func (s *NotificationService) checkIdempotency(ctx context.Context, notificationID string) error {
	key := cache.GetIdempotencyKey(notificationID)

	exists, err := s.cache.Exists(ctx, key)
	if err != nil {
		logger.Error("Failed to check idempotency",
			logger.Merge(
				logger.WithNotificationID(notificationID),
				logger.WithError(err),
			),
		)

		return nil // continue processing on cache error
	}

	if exists {
		return fmt.Errorf("notification already processed")
	}

	return nil
}

// mark notification as processed for idempotency
func (s *NotificationService) markAsProcessed(ctx context.Context, notificationID string) {
	key := cache.GetIdempotencyKey(notificationID)

	// store for 24 hours
	if err := s.cache.Set(ctx, key, "processed", 86400); err != nil {
		logger.Error("Failed to mark notification as processed",
			logger.Merge(
				logger.WithNotificationID(notificationID),
				logger.WithError(err),
			))
	}
}

// check if user has exceeded rate limit
func (s *NotificationService) checkRateLimit(ctx context.Context, userID string) error {
	key := cache.GetRateLimitKey(userID)

	// allow 100 notifications per minute per user
	allowed, err := s.cache.CheckRateLimit(ctx, key, 100, 60)
	if err != nil {
		logger.Error("Failed to check rate limit", logger.Merge(
			logger.WithUserID(userID),
			logger.WithError(err),
		))

		return nil // continue on cache error
	}

	if !allowed {
		return fmt.Errorf("rate limit exceeded for user %s", userID)
	}

	return nil
}

// validates device tokens(for testing purposes)
func (s *NotificationService) ValidateDeviceTokens(ctx context.Context, tokens []string) ([]*models.DeviceTokenValidation, error) {
	results := make([]*models.DeviceTokenValidation, 0, len(tokens))

	for _, token := range tokens {
		validation, err := s.fcmService.ValidateDeviceToken(ctx, token)
		if err != nil {
			logger.Error("Failed to validate device token",
				logger.Merge(logger.WithError(err), logger.Fields{
					"token": token,
				}))
			continue
		}
		results = append(results, validation)
	}

	return results, nil
}
