package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

// tests backoff duration calculation
func TestRetryServiceCalculateBackoff(t *testing.T) {
	service := NewRetryService(3, 1, 60, 2.0)

	testCases := []struct {
		name            string
		attemptCount    int
		expectedBackoff time.Duration
	}{
		{
			name:            "Initial attempt",
			attemptCount:    0,
			expectedBackoff: 1 * time.Second,
		},
		{
			name:            "Second attempt",
			attemptCount:    1,
			expectedBackoff: 2 * time.Second,
		},
		{
			name:            "Third attempt",
			attemptCount:    2,
			expectedBackoff: 4 * time.Second,
		},
		{
			name:            "Capped at max interval",
			attemptCount:    10,
			expectedBackoff: 60 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			backoff := service.CalculateBackoff(tc.attemptCount)
			if backoff != tc.expectedBackoff {
				t.Errorf("Expected backoff %v, got %v", tc.expectedBackoff, backoff)
			}
		})
	}
}

// tests retry with backoff logic
func TestRetryServiceRetryWithBackoff(t *testing.T) {
	service := NewRetryService(3, 1, 60, 2.0)

	t.Run("Success on first attempt", func(t *testing.T) {
		attempts := 0
		err := service.RetryWithBackoff(context.Background(), func() error {
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
		err := service.RetryWithBackoff(context.Background(), func() error {
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
		attempts := 0
		err := service.RetryWithBackoff(context.Background(), func() error {
			attempts++
			return errors.New("persistent error")
		})

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if attempts != 3 {
			t.Errorf("Expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("Context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		err := service.RetryWithBackoff(ctx, func() error {
			return errors.New("some error")
		})

		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	})
}
