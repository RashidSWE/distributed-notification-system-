package push

import (
	"errors"
	"sync"
	"time"

	"github.com/zjoart/distributed-notification-system/push-service/internal/models"
	"github.com/zjoart/distributed-notification-system/push-service/pkg/logger"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// state representation
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

type CircuitBreaker struct {
	maxRequests      uint32
	interval         time.Duration
	timeout          time.Duration
	failureThreshold uint32

	mutex           sync.Mutex
	state           State
	failures        uint32
	successes       uint32
	requests        uint32
	lastFailTime    time.Time
	lastStateChange time.Time
}

func NewCircuitBreaker(maxRequests, failureThreshold uint32, interval, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxRequests:      maxRequests,
		interval:         interval,
		timeout:          timeout,
		failureThreshold: failureThreshold,
		state:            StateClosed,
		lastStateChange:  time.Now(),
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	if err := cb.beforeCall(); err != nil {
		return err
	}

	err := fn()
	cb.afterCall(err)

	return err
}

func (cb *CircuitBreaker) beforeCall() error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()

	switch cb.state {
	case StateClosed:
		// reset counters if interval has passed
		if now.Sub(cb.lastStateChange) > cb.interval {
			cb.resetCounters()
		}
		return nil

	case StateOpen:
		// check if timeout has passed to transition to half-open
		if now.Sub(cb.lastFailTime) > cb.timeout {
			cb.setState(StateHalfOpen)
			return nil
		}
		return models.ErrCircuitBreakerOpen

	case StateHalfOpen:
		// allow limited requests in half-open state
		if cb.requests >= cb.maxRequests {
			return models.ErrCircuitBreakerOpen
		}
		cb.requests++
		return nil

	default:
		return errors.New("unknown circuit breaker state")
	}
}

func (cb *CircuitBreaker) afterCall(err error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
}

func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateClosed:
		cb.successes++

	case StateHalfOpen:
		cb.successes++
		// transition to closed if there are enough successful requests
		if cb.successes >= cb.maxRequests {
			cb.setState(StateClosed)
			logger.Info("Circuit breaker closed after successful recovery")
		}
	}
}

func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	switch cb.state {
	case StateClosed:
		// open the circuit if failure threshold has been exceeded
		if cb.failures >= cb.failureThreshold {
			cb.setState(StateOpen)
			logger.Warn("Circuit breaker opened due to failures", logger.Fields{
				"failures":  cb.failures,
				"threshold": cb.failureThreshold,
			})
		}

	case StateHalfOpen:
		// open the circuit if there are any failure in half-open state
		cb.setState(StateOpen)
		logger.Warn("Circuit breaker reopened after failure in half-open state")
	}
}

// changes circuit breaker state
func (cb *CircuitBreaker) setState(state State) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state
	cb.lastStateChange = time.Now()
	cb.resetCounters()

	logger.Info("Circuit breaker state changed", logger.Fields{
		"from": prev.String(),
		"to":   state.String(),
	})
}

// resets failure and success counters
func (cb *CircuitBreaker) resetCounters() {
	cb.failures = 0
	cb.successes = 0
	cb.requests = 0
}

// return current circuit breaker state
func (cb *CircuitBreaker) GetState() State {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	return cb.state
}

// return circuit breaker statistics
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	return map[string]interface{}{
		"state":             cb.state.String(),
		"failures":          cb.failures,
		"successes":         cb.successes,
		"requests":          cb.requests,
		"last_fail_time":    cb.lastFailTime,
		"last_state_change": cb.lastStateChange,
	}
}
