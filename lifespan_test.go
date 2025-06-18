package lifespan_test

import (
	"github.com/jharshman/lifespan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_Run(t *testing.T) {
	span := lifespan.Run(func(span *lifespan.LifeSpan) {
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

	// assert that span contains
	require.NotNil(t, span)
	assert.NotEmpty(t, span.UUID)
	assert.NotNil(t, span.Sig)
	assert.NotNil(t, span.Ack)
	assert.NotNil(t, span.Err)
	assert.NotNil(t, span.Ctx)
	assert.NotNil(t, span.Cancel)

	// close the span
	span.Close()
}
