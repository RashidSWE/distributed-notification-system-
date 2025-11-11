package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zjoart/distributed-notification-system/push-service/internal/models"
)

// tests backoff calculation
func TestRetryServiceCalculateBackoff(t *testing.T) {
	service := NewRetryService(3, 1, 60, 2.0)

	testCases := []struct {
		name        string
		attempt     int
		expectedMin time.Duration
		expectedMax time.Duration
	}{
		{
			name:        "First retry",
			attempt:     0,
			expectedMin: 1 * time.Second,
			expectedMax: 1 * time.Second,
		},
		{
			name:        "Second retry",
			attempt:     1,
			expectedMin: 2 * time.Second,
			expectedMax: 2 * time.Second,
		},
		{
			name:        "Third retry",
			attempt:     2,
			expectedMin: 4 * time.Second,
			expectedMax: 4 * time.Second,
		},
		{
			name:        "Capped at max",
			attempt:     10,
			expectedMin: 60 * time.Second,
			expectedMax: 60 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			backoff := service.CalculateBackoff(tc.attempt)
			if backoff < tc.expectedMin || backoff > tc.expectedMax {
				t.Errorf("Expected backoff between %v and %v, got %v", tc.expectedMin, tc.expectedMax, backoff)
			}
		})
	}
}

// tests retry decision logic
func TestRetryServiceShouldRetry(t *testing.T) {
	service := NewRetryService(3, 1, 60, 2.0)

	testCases := []struct {
		name         string
		attemptCount int
		shouldRetry  bool
	}{
		{
			name:         "First attempt - should retry",
			attemptCount: 0,
			shouldRetry:  true,
		},
		{
			name:         "Second attempt - should retry",
			attemptCount: 1,
			shouldRetry:  true,
		},
		{
			name:         "Third attempt - should retry",
			attemptCount: 2,
			shouldRetry:  true,
		},
		{
			name:         "Max attempts reached - should not retry",
			attemptCount: 3,
			shouldRetry:  false,
		},
		{
			name:         "Exceeded max attempts - should not retry",
			attemptCount: 4,
			shouldRetry:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			notification := &models.NotificationMessage{
				AttemptCount: tc.attemptCount,
			}

			result := service.ShouldRetry(notification)
			if result != tc.shouldRetry {
				t.Errorf("Expected shouldRetry=%v, got %v", tc.shouldRetry, result)
			}
		})
	}
}

// tests retry with backoff logic
func TestRetryServiceRetryWithBackoff(t *testing.T) {
	service := NewRetryService(3, 1, 60, 2.0)

	t.Run("Success on first attempt", func(t *testing.T) {
		attempts := 0
		err := service.RetryWithBackoff(context.Background(), 0, func() error {
			attempts++
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if attempts != 1 {
			t.Errorf("Expected 1 attempt, got %d", attempts)
		}
	})

	t.Run("Success on second attempt", func(t *testing.T) {
		attempts := 0
		err := service.RetryWithBackoff(context.Background(), 0, func() error {
			attempts++
			if attempts == 1 {
				return errors.New("temporary error") // return error on first attempt
			}
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if attempts != 2 {
			t.Errorf("Expected 2 attempts, got %d", attempts)
		}
	})

	t.Run("Max retries exceeded", func(t *testing.T) {
		err := service.RetryWithBackoff(context.Background(), 3, func() error {
			return errors.New("persistent error")
		})

		if err != models.ErrMaxRetriesExceeded {
			t.Errorf("Expected ErrMaxRetriesExceeded, got %v", err)
		}
	})

	t.Run("Context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		err := service.RetryWithBackoff(ctx, 0, func() error {
			return errors.New("some error")
		})

		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	})
}
