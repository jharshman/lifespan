package lifespan

import (
	"context"
	"errors"
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
	// GroupID identifies the group this LifeSpan belongs to, if any.
	GroupID string
	// Sig and Ack are the primary control channels. Write to Sig to signal to close, and read from Ack to acknowledge.
	Sig, Ack chan struct{}
	// ErrBus is an implementation of MessageBus[any T] which is shared across all Runnable implementations.
	ErrBus MessageBus[Error]
	// Logger is a unique log/slog.Logger for a lifespan. Lifespan's share a base log handler so that they may write to the same
	// underlying LogBus.
	Logger *slog.Logger
	// Context information for timeouts and cancels
	Ctx    context.Context
	Cancel context.CancelFunc
}

// Close will signal a runnable task to shutdown. If an acknowledgement is not given
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
// If groupID is empty, no group_id attribute will be added to the logger.
func Run(groupID string, logHandler slog.Handler, errBus MessageBus[Error], job func(span *LifeSpan)) (*LifeSpan, error) {

	// logHandler and errBus cannot be nil.
	if logHandler == nil {
		return nil, errors.New("nil logHandler")
	}

	if errBus == nil {
		return nil, errors.New("nil errBus")
	}

	id := uuid.New()

	// include job_id in logger created from logHandler
	l := slog.New(logHandler)
	l = l.With(slog.String("job_id", id.String()))

	if groupID != "" {
		l = l.With(slog.String("group_id", groupID))
	}

	ctx, cancel := context.WithCancel(context.Background())
	span := &LifeSpan{
		UUID:    id.String(),
		GroupID: groupID,
		Sig:     make(chan struct{}, 1),
		Ack:     make(chan struct{}, 1),
		ErrBus:  errBus,
		Logger:  l,
		Ctx:     ctx,
		Cancel:  cancel,
	}

	go func() {
		defer close(span.Ack)
		defer cancel()
		job(span)
	}()

	return span, nil
}

// Error shortcuts publishing to the ErrBus and inserts the JobID, GroupID, and timestamp into the Error.
func (span *LifeSpan) Error(err error) {
	e := Error{
		JobID:     span.UUID,
		GroupID:   span.GroupID,
		Error:     err,
		Timestamp: time.Now().UTC(),
	}
	span.ErrBus.Publish(e)
}
