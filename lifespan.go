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
	UUID     string
	Sig, Ack chan struct{}
	Err      chan error
	Ctx      context.Context
	Cancel   context.CancelFunc
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
func Run(job func(span *LifeSpan)) (span *LifeSpan) {
	ctx, cancel := context.WithCancel(context.Background())
	id := uuid.New()

	span = &LifeSpan{
		UUID:   id.String(),
		Sig:    make(chan struct{}, 1),
		Ack:    make(chan struct{}, 1),
		Err:    make(chan error, 1),
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
