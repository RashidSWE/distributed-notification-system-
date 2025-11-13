package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zjoart/distributed-notification-system/push-service/internal/cache"
	"github.com/zjoart/distributed-notification-system/push-service/internal/config"
	"github.com/zjoart/distributed-notification-system/push-service/internal/models"
	"github.com/zjoart/distributed-notification-system/push-service/internal/push"
	"github.com/zjoart/distributed-notification-system/push-service/internal/template"
	"github.com/zjoart/distributed-notification-system/push-service/pkg/logger"
)

type NotificationService struct {
	fcmService     *push.FCMService
	retryService   *RetryService
	cache          *cache.RedisCache
	rateLimit      config.RateLimitConfig
	queue          QueuePublisher
	templateClient *template.Client
}

type QueuePublisher interface {
	PublishStatus(ctx context.Context, statusMsg *models.NotificationStatusMessage) error
}

func NewNotificationService(
	fcmService *push.FCMService,
	retryService *RetryService,
	cache *cache.RedisCache,
	rateLimit config.RateLimitConfig,
	queue QueuePublisher,
	templateClient *template.Client,
) *NotificationService {
	return &NotificationService{
		fcmService:     fcmService,
		retryService:   retryService,
		cache:          cache,
		rateLimit:      rateLimit,
		queue:          queue,
		templateClient: templateClient,
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

		// publish failed status for rate limit
		s.publishStatus(ctx, msg, nil, models.NotificationStatusFailed, "Rate limit exceeded", 0, 0)
		return err
	}

	if err := msg.Validate(); err != nil {
		logger.Error("Invalid notification message",
			logger.Merge(loggerDetails, logger.WithError(err)),
		)
		// publish failed status for validation errors
		s.publishStatus(ctx, msg, nil, models.NotificationStatusFailed, fmt.Sprintf("Validation failed: %s", err.Error()), 0, 0)
		return err
	}

	if err := s.checkIdempotency(ctx, msg.ID); err != nil {
		logger.Warn("Duplicate notification detected", loggerDetails)

		return nil // return nil to acknowledge the message
	}

	logger.Info("Processing notification", logger.Merge(loggerDetails, logger.Fields{
		"device_count": len(msg.DeviceTokens),
	}))

	notification, err := s.prepareNotification(ctx, msg)
	if err != nil {
		logger.Error("Failed to prepare notification", logger.Merge(loggerDetails,
			logger.WithError(
				err,
			)))
		// publish failed status for preparation errors
		s.publishStatus(ctx, msg, nil, models.NotificationStatusFailed, fmt.Sprintf("Failed to prepare notification: %s", err.Error()), 0, 0)
		return err
	}

	results, err := s.sendNotification(ctx, msg, notification)
	if err != nil {
		logger.Error("Failed to send notification", logger.Merge(loggerDetails,
			logger.WithError(
				err,
			)))

		if results == nil {
			results = make([]*models.NotificationResult, 0)
		}

		// publish failed status
		s.publishStatus(ctx, msg, results, models.NotificationStatusFailed, err.Error(), 0, len(results))
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

	// determine final status
	finalStatus := models.NotificationStatusDelivered
	statusMessage := "Notification delivered successfully"
	if successCount == 0 {
		finalStatus = models.NotificationStatusFailed
		statusMessage = "All notifications failed to deliver"
	} else if failedCount > 0 {
		statusMessage = fmt.Sprintf("Partially delivered: %d succeeded, %d failed", successCount, failedCount)
	}

	// publish status to status queue
	s.publishStatus(ctx, msg, results, finalStatus, statusMessage, successCount, failedCount)

	// mark as processed for idempotency
	s.markAsProcessed(ctx, msg.ID)

	return nil
}

// prepare the notification content
func (s *NotificationService) prepareNotification(ctx context.Context, msg *models.NotificationMessage) (*models.PushNotification, error) {
	notification := &models.PushNotification{
		Title:    msg.Title,
		Body:     msg.Body,
		ImageURL: msg.ImageURL,
		Link:     msg.Link,
		Data:     msg.Data,
		Priority: msg.Priority,
	}

	logDetails := logger.Merge(

		logger.WithNotificationID(msg.ID),
		logger.Fields{
			"template_code": msg.TemplateCode,
		},
	)

	// fetch template if template code is provided
	if msg.TemplateCode != "" {
		logger.Info("Fetching template for notification", logDetails)

		tmpl, err := s.templateClient.GetPushTemplate(ctx, msg.TemplateCode)
		if err != nil {
			logger.Error("Failed to fetch template", logger.Merge(
				logger.WithError(err),
				logDetails,
			))
			return nil, fmt.Errorf("failed to fetch template: %w", err)
		}

		// apply template values with variable replacement
		notification.Title = s.replaceVariables(tmpl.Title, msg.Variables)
		notification.Body = s.replaceVariables(tmpl.Body, msg.Variables)

		if tmpl.ImageURL != "" {
			notification.ImageURL = tmpl.ImageURL
		}

		// merge template data with message data
		if tmpl.Data != nil {
			if notification.Data == nil {
				notification.Data = make(map[string]interface{})
			}
			for key, value := range tmpl.Data {
				notification.Data[key] = value
			}
		}

		// override with message-specific data
		if msg.Data != nil {
			for key, value := range msg.Data {
				notification.Data[key] = value
			}
		}

		logger.Info("Template applied successfully", logger.Merge(logDetails, logger.Fields{
			"title": notification.Title,
		}))
	}

	return notification, nil
}

// replaceVariables replaces {{variable}} placeholders with actual values
func (s *NotificationService) replaceVariables(text string, variables map[string]string) string {
	if variables == nil {
		return text
	}

	result := text
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
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

	err := s.retryService.RetryWithBackoff(ctx, func() error {
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

	// use configured rate limit values
	allowed, err := s.cache.CheckRateLimit(ctx, key, int64(s.rateLimit.Requests), s.rateLimit.Window)
	if err != nil {
		logger.Error("Failed to check rate limit", logger.Merge(
			logger.WithUserID(userID),
			logger.WithError(err),
		))

		return nil // continue on cache error
	}

	if !allowed {
		return models.ErrRateLimitExceeded
	}

	return nil
}

// publishes notification status to the status queue
func (s *NotificationService) publishStatus(ctx context.Context, msg *models.NotificationMessage, results []*models.NotificationResult, status models.NotificationStatusEnum, message string, successCount, failedCount int) {
	statusMsg := &models.NotificationStatusMessage{
		RequestID:      msg.RequestID,
		NotificationID: msg.ID,
		UserID:         msg.UserID,
		Status:         status,
		Message:        message,
		SentCount:      successCount,
		FailedCount:    failedCount,
		Results:        results,
		Timestamp:      time.Now(),
		CorrelationID:  msg.CorrelationID,
	}

	if err := s.queue.PublishStatus(ctx, statusMsg); err != nil {
		logger.Error("Failed to publish status to queue",
			logger.Merge(
				logger.WithNotificationID(msg.ID),
				logger.WithUserID(msg.UserID),
				logger.WithError(err),
			))
	}
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
