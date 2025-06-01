package lifespan

import (
	"context"
	"log/slog"
	"time"
)

type Runnable interface {
  Run(span *LifeSpan)
}

type LifeSpan struct {
  Sig, Ack chan struct {}
  Err chan error
  Ctx context.Context
  Cancel context.CancelFunc
}

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

func Run(f func(span *LifeSpan)) (span *LifeSpan) {
  ctx, cancel := context.WithCancel(context.Background())

  span = &LifeSpan{
    Sig: make(chan struct{}, 1),
    Ack: make(chan struct{}, 1),
    Err: make(chan error, 1),
    Ctx: ctx,
    Cancel: cancel,
  }

  go func() {
    defer close(span.Ack)
    defer cancel()
    f(span)
  }()

  return
}

