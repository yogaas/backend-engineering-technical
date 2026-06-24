package external

import (
	"errors"
	"sync"
	"time"
)

var ErrCircuitOpen = errors.New("circuit breaker terbuka, layanan pihak ketiga sedang dihindari sementara")

type state int

const (
	closed state = iota
	open
	halfOpen
)

type CircuitBreaker struct {
	mu               sync.Mutex
	state            state
	failureCount     int
	failureThreshold int
	openUntil        time.Time
	openDuration     time.Duration
}

func NewCircuitBreaker(failureThreshold int, openDuration time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		openDuration:     openDuration,
		state:            closed,
	}
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case open:
		if time.Now().After(cb.openUntil) {
			cb.state = halfOpen
			return true
		}
		return false
	default:
		return true
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount = 0
	cb.state = closed
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount++
	if cb.state == halfOpen || cb.failureCount >= cb.failureThreshold {
		cb.state = open
		cb.openUntil = time.Now().Add(cb.openDuration)
	}
}
