package models

import "errors"

var (
	// Notification errors
	ErrInvalidNotificationID     = errors.New("invalid notification ID")
	ErrInvalidUserID             = errors.New("invalid user ID")
	ErrNoDeviceTokens            = errors.New("no device tokens provided")
	ErrEmptyNotificationContent  = errors.New("notification content cannot be empty")
	ErrInvalidDeviceToken        = errors.New("invalid device token")
	ErrTemplateNotFound          = errors.New("template not found")
	ErrTemplateVariableMissing   = errors.New("required template variable missing")
	ErrInvalidRequestID          = errors.New("invalid request ID")
	ErrInvalidNotificationStatus = errors.New("invalid notification status")

	// User errors
	ErrInvalidUserName = errors.New("invalid user name")
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPassword = errors.New("invalid password")

	// Service errors
	ErrCircuitBreakerOpen    = errors.New("circuit breaker is open")
	ErrMaxRetriesExceeded    = errors.New("max retry attempts exceeded")
	ErrFCMServiceUnavailable = errors.New("FCM service unavailable")
	ErrInvalidFCMResponse    = errors.New("invalid FCM response")
	ErrRateLimitExceeded     = errors.New("rate limit exceeded")

	// Database errors
	ErrDatabaseConnection = errors.New("database connection error")
	ErrRecordNotFound     = errors.New("record not found")
	ErrDuplicateRecord    = errors.New("duplicate record")

	// Cache errors
	ErrCacheConnection = errors.New("cache connection error")
	ErrCacheMiss       = errors.New("cache miss")

	// Queue errors
	ErrQueueConnection      = errors.New("queue connection error")
	ErrMessagePublishFailed = errors.New("failed to publish message")
	ErrMessageConsumeFailed = errors.New("failed to consume message")
	ErrInvalidMessageFormat = errors.New("invalid message format")
)
