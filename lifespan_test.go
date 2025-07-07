package lifespan_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jharshman/lifespan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Run(t *testing.T) {

	span, _ := lifespan.Run(context.Background(), func(ctx context.Context, span *lifespan.LifeSpan) {
		span.Logger.Info("testing")
	LOOP:
		for {
			select {
			case <-ctx.Done():
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
	assert.NotNil(t, span.Sig)
	assert.NotNil(t, span.Ack)
	assert.NotNil(t, span.ErrBus)

	// close the span
	span.Close()

}

func Test_RunWithErrorBus(t *testing.T) {

	jobfunc := func(ctx context.Context, span *lifespan.LifeSpan) {
		span.Logger.Info("starting job")
	LOOP:
		for {
			select {
			case <-ctx.Done():
				break LOOP
			case <-span.Sig:
				span.Error(ctx, errors.New("test error"))
			}
		}
		span.Ack <- struct{}{}
	}

	ctx, cancel := context.WithCancel(context.Background())
	span1, _ := lifespan.Run(ctx, jobfunc)
	span2, _ := lifespan.Run(ctx, jobfunc)

	errorCount := &atomic.Int32{}
	span3, _ := lifespan.Run(ctx, func(ctx context.Context, span *lifespan.LifeSpan) {
		sub := lifespan.DefaultCentralErrorBus.Subscribe()
	LOOP:
		for {
			select {
			case <-span.Sig:
				break LOOP
			case msg := <-sub:
				fmt.Println(msg)
				errorCount.Add(1)
			}
		}
	})

	// trigger write for span1 and span2 to error bus
	for i := 0; i < 5; i++ {
		span1.Sig <- struct{}{}
		span2.Sig <- struct{}{}
	}

LOOP:
	for {
		select {
		case <-time.After(time.Second * 3):
			break LOOP // kill this check after 3 seconds
		default:
			if errorCount.Load() >= 10 {
				break LOOP
			}
		}
	}

	// kill span1 and span2 with cancel function
	cancel()
	// kill span3
	span3.Close()

	// close central error bus
	lifespan.DefaultCentralErrorBus.Close()

	assert.Equal(t, int32(10), errorCount.Load())

}
