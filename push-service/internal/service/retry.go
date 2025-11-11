package service

import (
	"context"
	"math"
	"time"

	"github.com/zjoart/distributed-notification-system/push-service/internal/models"
	"github.com/zjoart/distributed-notification-system/push-service/pkg/logger"
)

type RetryService struct {
	maxAttempts     int
	initialInterval time.Duration
	maxInterval     time.Duration
	multiplier      float64
}

func NewRetryService(maxAttempts, initialInterval, maxInterval int, multiplier float64) *RetryService {
	return &RetryService{
		maxAttempts:     maxAttempts,
		initialInterval: time.Duration(initialInterval) * time.Second,
		maxInterval:     time.Duration(maxInterval) * time.Second,
		multiplier:      multiplier,
	}
}

// calculates the backoff duration for retry attempt
func (r *RetryService) CalculateBackoff(attemptCount int) time.Duration {
	if attemptCount <= 0 {
		return r.initialInterval
	}

	// calculates exponential backoff: initialInterval * multiplier^attemptCount
	backoff := float64(r.initialInterval) * math.Pow(r.multiplier, float64(attemptCount))

	// cap at maximum interval
	if backoff > float64(r.maxInterval) {
		backoff = float64(r.maxInterval)
	}

	return time.Duration(backoff)
}

// determines if a notification should be retried
func (r *RetryService) ShouldRetry(notification *models.NotificationMessage) bool {
	return notification.AttemptCount < r.maxAttempts
}

// retries a function with exponential backoff
func (r *RetryService) RetryWithBackoff(ctx context.Context, attemptCount int, fn func() error) error {
	if attemptCount >= r.maxAttempts {
		return models.ErrMaxRetriesExceeded
	}

	var lastErr error

	for attempt := attemptCount; attempt < r.maxAttempts; attempt++ {

		// execute
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// check if we should continue retrying
		if attempt < r.maxAttempts-1 {
			backoff := r.CalculateBackoff(attempt)

			logger.Info("Retrying after backoff", logger.Fields{
				"attempt": attempt + 1,
				"backoff": backoff.String(),
				"error":   err.Error(),
			})

			// wait for backoff period or context cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				continue
			}
		}
	}

	return lastErr
}

// returns the maximum number of retry attempts
func (r *RetryService) GetMaxAttempts() int {
	return r.maxAttempts
}
