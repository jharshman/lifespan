package lifespan

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

const (
	defaultBufferSize = 1024
	jobIDKey          = "job_id"
	groupIDKey        = "group_id"
)

// Runnable defines the behavior of a runnable task.
type Runnable interface {
	Run(ctx context.Context, span *LifeSpan)
}

// LifeSpan holds the communication channels and context for a runnable task.
type LifeSpan struct {
	// Sig and Ack are the primary control channels. Write to Sig to signal to close, and read from Ack to acknowledge.
	Sig, Ack chan struct{}
	//
	ErrBus chan Error
	// Default logger with extra context injected via Run.
	Logger *slog.Logger
}

// Close will signal a runnable task to shutdown. If an acknowledgement is not given
// by the runnable task after 3 seconds, Close will log a warning but otherwise
// leave the task to handle cancellation according to its own implementation.
func (span *LifeSpan) Close() {
	select {
	case span.Sig <- struct{}{}:
		select {
		case <-span.Ack:
			return
		case <-time.After(3 * time.Second):
			slog.Warn("timeout waiting for acknowledgement")
		}
	default:
		slog.Warn("unable to send signal")
	}
}

// Run runs the passed in job and returns a pointer to a LifeSpan.
// If groupID is empty, no group_id attribute will be added to the logger.
func Run(ctx context.Context, job func(ctx context.Context, span *LifeSpan)) (*LifeSpan, error) {

	// if the context does not contain a job_id then create and set one.
	if _, ok := ctx.Value(jobIDKey).(string); !ok {
		ctx = context.WithValue(ctx, jobIDKey, uuid.New().String())
	}

	// create a unique channel for the LifeSpan's errors and register it with the DefaultCentralErrorBus.
	errChan := make(chan Error, defaultBufferSize)
	DefaultCentralErrorBus.Register(errChan)

	// The context should contain the job_id and possibly the group_id and is the source of truth for these values.
	// Create a new Logger from the logHandler and set the job_id and group_id attributes.
	// This provides a fallback to ensure that we have these values in logs regardless if the user chooses to log with context.
	l := slog.Default().With(
		slog.String(jobIDKey, JobIDFromContext(ctx)),
		slog.String(groupIDKey, GroupIDFromContext(ctx)),
	)

	span := &LifeSpan{
		Sig:    make(chan struct{}, 1),
		Ack:    make(chan struct{}, 1),
		ErrBus: errChan,
		Logger: l,
	}

	go func() {
		defer close(span.Ack)
		defer close(span.ErrBus)
		job(ctx, span)
	}()

	return span, nil
}

// Error shortcuts publishing to the ErrBus and inserts the JobID, GroupID, and timestamp into the Error.
func (span *LifeSpan) Error(ctx context.Context, err error) {
	e := Error{
		Error:     err,
		Timestamp: time.Now().UTC(),
	}
	if jid, ok := ctx.Value(jobIDKey).(string); ok {
		e.JobID = jid
	}
	if gid, ok := ctx.Value(groupIDKey).(string); ok {
		e.GroupID = gid
	}

	span.ErrBus <- e
}

func JobIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(jobIDKey).(string); ok {
		return v
	}
	return ""
}

func GroupIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(groupIDKey).(string); ok {
		return v
	}
	return ""
}
