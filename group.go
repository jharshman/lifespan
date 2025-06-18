package lifespan

import (
	"context"

	"github.com/google/uuid"
)

// Group defines a grouping of Runnable jobs and LifeSpans.
type Group struct {
	UUID   string
	Jobs   []Runnable
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
func (group *Group) Start() {
	for _, job := range group.Jobs {
		span := Run(func(span *LifeSpan) {
			job.Run(span)
		})
		group.Spans = append(group.Spans, span)
	}
}

// Close will cancel the Group and range over available spans calling each span's Close Method.
func (group *Group) Close() {
	group.Cancel()
	for _, span := range group.Spans {
		span.Close()
	}
}
