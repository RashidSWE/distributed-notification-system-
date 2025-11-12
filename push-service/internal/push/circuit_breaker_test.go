package push

import (
	"errors"
	"testing"
	"time"
)

// test circuit breaker in closed state
func TestCircuitBreakerClosed(t *testing.T) {
	cb := NewCircuitBreaker(3, 5, 60*time.Second, 30*time.Second)

	// should allow calls in closed state
	for i := 0; i < 3; i++ {
		err := cb.Call(func() error {
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error in closed state, got %v", err)
		}
	}

	if cb.GetState() != StateClosed {
		t.Errorf("Expected state to be closed, got %s", cb.GetState().String())
	}
}

// test circuit breaker opens after failures
func TestCircuitBreakerOpensOnFailures(t *testing.T) {
	cb := NewCircuitBreaker(3, 3, 60*time.Second, 1*time.Second)

	// fail threshold times
	for i := 0; i < 3; i++ {
		cb.Call(func() error {
			return errors.New("test error")
		})
	}

	// should be open now
	if cb.GetState() != StateOpen {
		t.Errorf("Expected state to be open after failures, got %s", cb.GetState().String())
	}

	// should reject calls when open
	err := cb.Call(func() error {
		return nil
	})

	if err == nil {
		t.Error("Expected circuit breaker to reject call when open")
	}
}

// tests circuit breaker half-open state
func TestCircuitBreakerHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(3, 3, 60*time.Second, 100*time.Millisecond)

	// open the circuit
	for i := 0; i < 3; i++ {
		cb.Call(func() error {
			return errors.New("test error")
		})
	}

	// wait for timeout
	time.Sleep(150 * time.Millisecond)

	// should transition to half-open and allow limited requests
	err := cb.Call(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected call to succeed in half-open state, got %v", err)
	}

	state := cb.GetState()
	if state != StateHalfOpen && state != StateClosed {
		t.Errorf("Expected state to be half-open or closed, got %s", state.String())
	}
}

// test circuit breaker recovery to closed state
func TestCircuitBreakerRecovery(t *testing.T) {
	cb := NewCircuitBreaker(2, 3, 60*time.Second, 100*time.Millisecond)

	// open the circuit
	for i := 0; i < 3; i++ {
		cb.Call(func() error {
			return errors.New("test error")
		})
	}

	// wait for timeout to enter half-open
	time.Sleep(150 * time.Millisecond)

	// make successful calls to close the circuit
	for i := 0; i < 2; i++ {
		err := cb.Call(func() error {
			return nil
		})

		if err != nil {
			t.Errorf("Expected successful call in half-open state, got %v", err)
		}
	}

	// should be closed now
	if cb.GetState() != StateClosed {
		t.Errorf("Expected state to be closed after recovery, got %s", cb.GetState().String())
	}
}

// test circuit breaker statistics
func TestCircuitBreakerStats(t *testing.T) {
	cb := NewCircuitBreaker(3, 5, 60*time.Second, 30*time.Second)

	// make some calls
	cb.Call(func() error { return nil })
	cb.Call(func() error { return errors.New("test error") })
	cb.Call(func() error { return nil })

	stats := cb.GetStats()

	if stats["state"] != "closed" {
		t.Errorf("Expected state 'closed', got %v", stats["state"])
	}

	if stats["failures"].(uint32) != 1 {
		t.Errorf("Expected 1 failure, got %v", stats["failures"])
	}

	if stats["successes"].(uint32) != 2 {
		t.Errorf("Expected 2 successes, got %v", stats["successes"])
	}
}

// test state string representation
func TestCircuitBreakerStateString(t *testing.T) {
	testCases := []struct {
		state    State
		expected string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			if tc.state.String() != tc.expected {
				t.Errorf("Expected state string '%s', got '%s'", tc.expected, tc.state.String())
			}
		})
	}
}
