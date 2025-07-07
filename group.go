package lifespan

import (
	"context"

	"github.com/google/uuid"
)

// Group defines a grouping of Runnable jobs and LifeSpans.
type Group struct {
	// UUID identifies a Group of Runnable. Useful for attributing logs and errors to a group.
	UUID string
	// Jobs an array of Runnable
	Jobs []Runnable
	// Spans is a map of LifeSpans keyed by the LifeSpan's UUID.
	Spans map[string]*LifeSpan
}

// NewGroup returns a pointer to a *Group holding the Runnable jobs.
func NewGroup(jobs ...Runnable) *Group {
	id := uuid.New()
	return &Group{
		UUID:  id.String(),
		Jobs:  jobs,
		Spans: make(map[string]*LifeSpan, len(jobs)),
	}
}

// Start executes the group of Jobs, storing each Job's LifeSpan in the Group structure.
func (group *Group) Start() error {

	// base context contains group_id
	baseCtx := context.Background()
	baseCtx = context.WithValue(baseCtx, groupIDKey, group.UUID)

	for _, job := range group.Jobs {
		// build context per job containing job_id
		id := uuid.New().String()
		ctx := context.WithValue(baseCtx, jobIDKey, id)
		span, _ := Run(ctx, func(ctx context.Context, span *LifeSpan) {
			job.Run(ctx, span)
		})
		group.Spans[id] = span
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
	return group.Spans[uuid]
}
