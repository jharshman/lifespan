package lifespan_test

import (
	"errors"
	"github.com/jharshman/lifespan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_Run(t *testing.T) {

	span := lifespan.Run(nil, nil, func(span *lifespan.LifeSpan) {
		t.Logf("started job: %s", span.UUID)
	LOOP:
		for {
			select {
			case <-span.Ctx.Done():
				break LOOP
			case <-span.Sig:
				break LOOP
			default:
			}
		}
		span.Ack <- struct{}{}
	})

	// assert that span contains...
	require.NotNil(t, span)
	assert.NotEmpty(t, span.UUID)
	assert.NotNil(t, span.Sig)
	assert.NotNil(t, span.Ack)
	assert.Nil(t, span.ErrBus)
	assert.Nil(t, span.LogBus)
	assert.NotNil(t, span.Ctx)
	assert.NotNil(t, span.Cancel)

	// close the span
	span.Close()
}

func Test_RunWithErrorBus(t *testing.T) {

	// create a Message Bus for errors
	bus := &lifespan.ErrorBus{
		Bus: make(chan lifespan.Error, 10),
	}

	span := lifespan.Run(nil, bus, func(span *lifespan.LifeSpan) {
		t.Logf("started job: %s", span.UUID)
	LOOP:
		for {
			select {
			case <-span.Ctx.Done():
				break LOOP
			case <-span.Sig:
				break LOOP
			default:
			}
		}
		// test publish error
		span.ErrBus.Publish(lifespan.Error{
			JobID: "123-456-789",
			Error: errors.New("testing 123"),
		})
		span.Ack <- struct{}{}
	})

	// assert that span contains...
	require.NotNil(t, span)
	assert.NotEmpty(t, span.UUID)
	assert.NotNil(t, span.Sig)
	assert.NotNil(t, span.Ack)
	assert.NotNil(t, span.ErrBus)
	assert.NotNil(t, span.Ctx)
	assert.NotNil(t, span.Cancel)

	// subscribe to errBus
	e := bus.Subscribe()
	assert.NotNil(t, e)

	// close the span
	// this also tests the errBus
	span.Close()

	// read error from errBus
	errVal := <-e
	assert.NotNil(t, errVal)
	assert.Equal(t, "123-456-789", errVal.JobID)
}

func Test_RunWithMoreJobsAndErrors(t *testing.T) {
	// create a Message Bus for errors
	bus := &lifespan.ErrorBus{
		Bus: make(chan lifespan.Error, 10),
	}

	span1 := lifespan.Run(nil, bus, func(span *lifespan.LifeSpan) {
		t.Logf("started job: %s", span.UUID)
	LOOP:
		for {
			select {
			case <-span.Ctx.Done():
				break LOOP
			case <-span.Sig:
				span.ErrBus.Publish(lifespan.Error{
					JobID: span.UUID,
					Error: errors.New("testing 123"),
				})
				break LOOP
			default:
			}
		}
		span.Ack <- struct{}{}
	})

	span2 := lifespan.Run(nil, bus, func(span *lifespan.LifeSpan) {
		t.Logf("started job: %s", span.UUID)
	LOOP:
		for {
			select {
			case <-span.Ctx.Done():
				break LOOP
			case <-span.Sig:
				span.ErrBus.Publish(lifespan.Error{
					JobID: span.UUID,
					Error: errors.New("testing 456"),
				})
				break LOOP
			default:
			}
		}
		span.Ack <- struct{}{}
	})

	// clean up jobs
	// This will also trigger a write to the errBus for span1 and span2.
	span1.Close()
	span2.Close()

	// span1 and span2 are done, so we can close the errBus to prevent deadlock when looping over the values in the channel below.
	bus.Close()

	// read remaining data in buffered errBus channel.
	aggregateErrors := bus.Subscribe()
	errCount := 0
	for val := range aggregateErrors {
		errCount++
		assert.NotNil(t, val)
		assert.Error(t, val.Error)
	}
	assert.Equal(t, 2, errCount)
}
