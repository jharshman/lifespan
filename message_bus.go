package lifespan

import (
	"sync"
	"time"
)

// Error is a message that can be sent via a lifespan MessageBus.
// This type contains a standard error with additional metadata like the lifespan's UUID.
type Error struct {
	JobID     string
	GroupID   string
	Error     error
	Timestamp time.Time
}

// MessageBus defines behavior for a generic message bus.
// The implementations within the lifespan package provide
// a MessageBus for Errors and a MessageBus for Logs.
type MessageBus[T any] interface {
	Register(ch <-chan T)
	Publish(msg T)
	Subscribe() <-chan T
	Close()
}

// CentralMessageBus ...
type CentralMessageBus[T any] struct {
	mu     sync.Mutex
	wg     sync.WaitGroup
	closed bool
	bus    chan T
}

var DefaultCentralErrorBus = NewCentralMessageBus[Error](defaultBufferSize)

func NewCentralMessageBus[T any](bufferSize int64) *CentralMessageBus[T] {
	return &CentralMessageBus[T]{
		bus: make(chan T, bufferSize),
	}
}

func (cb *CentralMessageBus[T]) Register(ch <-chan T) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// check if the message bus is closed
	// don't accept new registration if the bus is closed.
	if cb.closed {
		panic("CentralMessageBus[T] is already closed")
	}

	// for the ch parameter, which is a receive only channel of type T,
	// spawn a goroutine to write into the central message bus
	cb.wg.Add(1)
	go func() {
		defer cb.wg.Done()
		// the loop ends when ch is closed.
		for msg := range ch {
			cb.bus <- msg
		}
	}()
}

// Publish writes the Error to the bus.
func (cb *CentralMessageBus[T]) Publish(msg T) {
	select {
	case cb.bus <- msg:
	default:
		// todo: record dropped messages metric
	}
}

// Subscribe returns a receive channel for the ErrorBus implementation.
func (cb *CentralMessageBus[T]) Subscribe() <-chan T {
	return cb.bus
}

// Close will close the channel contained within ErrorBus.
func (cb *CentralMessageBus[T]) Close() {
	cb.mu.Lock()

	// check if already closed
	if cb.closed {
		cb.mu.Unlock()
		return
	}

	// set close to true
	// this stops additional registrations
	cb.closed = true
	cb.mu.Unlock()

	// wait for all publishers to close first.
	cb.wg.Wait()

	close(cb.bus)
}
