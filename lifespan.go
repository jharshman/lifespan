package lifespan

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// Runnable defines the behavior of a runnable task.
type Runnable interface {
	Run(span *LifeSpan)
}

// LifeSpan holds the communication channels and context for a runnable task.
type LifeSpan struct {
	// UUID identifies the Lifespan for a Job. Useful for attributing logs and errors to jobs.
	UUID string
	// Sig and Ack are the primary control channels. Write to Sig to signal to close, and read from Ack to acknowledge.
	Sig, Ack chan struct{}
	// ErrBus is an implementation of MessageBus[any T] which is shared across all Runnable implementations.
	ErrBus MessageBus[Error]
	Ctx    context.Context
	Cancel context.CancelFunc
}

// Close will signal a runnable task to shutdown. If an acknoledgement is not given
// by the runnable task after 3 seconds, Close will move on.
func (span *LifeSpan) Close() {
	select {
	case span.Sig <- struct{}{}:
		select {
		case <-span.Ack:
			return
		case <-time.After(3 * time.Second):
			slog.Warn("timeout waiting for ack")
		}
	default:
	}
}

// Run runs the passed in job and returns a pointer to a LifeSpan.
func Run(errBus MessageBus[Error], job func(span *LifeSpan)) (span *LifeSpan) {
	ctx, cancel := context.WithCancel(context.Background())
	id := uuid.New()

	span = &LifeSpan{
		UUID:   id.String(),
		Sig:    make(chan struct{}, 1),
		Ack:    make(chan struct{}, 1),
		ErrBus: errBus,
		Ctx:    ctx,
		Cancel: cancel,
	}

	go func() {
		defer close(span.Ack)
		defer cancel()
		job(span)
	}()

	return
}
