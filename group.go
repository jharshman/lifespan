package lifespan

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// Group defines a grouping of Runnable jobs and LifeSpans.
type Group struct {
	// UUID identifies a Group of Runnable. Useful for attributing logs and errors to a group.
	UUID string
	// Jobs an array of Runnable
	Jobs []Runnable
	// Array of *LifeSpan
	Spans  []*LifeSpan
	Ctx    context.Context
	Cancel context.CancelFunc
}

// NewGroup returns a pointer to a *Group holding the Runnable jobs.
func NewGroup(jobs ...Runnable) *Group {
	ctx, cancel := context.WithCancel(context.Background())
	id := uuid.New()
	return &Group{
		UUID:   id.String(),
		Jobs:   jobs,
		Ctx:    ctx,
		Cancel: cancel,
	}
}

// Start executes the group of Jobs, storing each Job's LifeSpan in the Group structure.
func (group *Group) Start(logHandler *Logger, errBus *ErrorBus) error {
	if logHandler == nil {
		return errors.New("nil logHandler")
	}
	if errBus == nil {
		return errors.New("nil errBus")
	}
	for _, job := range group.Jobs {
		span, _ := Run(group.UUID, logHandler, errBus, func(span *LifeSpan) {
			job.Run(span)
		})
		group.Spans = append(group.Spans, span)
	}
	return nil
}

// Close will range over available spans calling each span's Close Method.
func (group *Group) Close() {
	for _, span := range group.Spans {
		span.Close()
	}
}

// GetLifeSpanByID returns a pointer to the LifeSpan associated with the given uuid.
// returns nil if non exists.
func (group *Group) GetLifeSpanByID(uuid string) *LifeSpan {
	for _, span := range group.Spans {
		if span.UUID == uuid {
			return span
		}
	}
	return nil
}
