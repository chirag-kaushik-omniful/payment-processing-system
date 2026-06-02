package circuitbreaker

import (
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

type Breaker struct {
	mu            sync.Mutex
	state         State
	failures      int
	threshold     int
	openUntil     time.Time
	resetTimeout  time.Duration
}

func New(threshold int, resetTimeout time.Duration) *Breaker {
	return &Breaker{threshold: threshold, resetTimeout: resetTimeout, state: StateClosed}
}

func (b *Breaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state == StateOpen {
		if time.Now().After(b.openUntil) {
			b.state = StateHalfOpen
			return true
		}
		return false
	}
	return true
}

func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.state = StateClosed
}

func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	if b.failures >= b.threshold {
		b.state = StateOpen
		b.openUntil = time.Now().Add(b.resetTimeout)
	}
}
