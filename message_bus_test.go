package lifespan_test

import (
	"testing"

	"github.com/jharshman/lifespan"
	"github.com/stretchr/testify/require"
)

func TestNewErrorBus(t *testing.T) {
	errBus := lifespan.NewErrorBus(0) // bsize of zero is invalid but will default to defaultBufferSize
	defer errBus.Close()

	require.NotNil(t, errBus)        // NewErrorBus should never return a nil bus
	errBus.Publish(lifespan.Error{}) // write to errBus empty Error
	e := <-errBus.Subscribe()        // subscribe to bus & consume a message
	t.Logf("%v", e)
}

func TestNewLogBus(t *testing.T) {
	logBus := lifespan.NewLogBus(0) // bsize of zero is invalid but will default to defaultBufferSize
	defer logBus.Close()

	require.NotNil(t, logBus)      // NewErrorBus should never return a nil bus
	logBus.Publish(lifespan.Log{}) // write to errBus empty Log
	e := <-logBus.Subscribe()      // subscribe to bus & consume a message
	t.Logf("%v", e)
}
