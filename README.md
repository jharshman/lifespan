# Lifespan
[![Go Report Card](https://goreportcard.com/badge/github.com/jharshman/lifespan)](https://goreportcard.com/report/github.com/jharshman/lifespan)
[![Go Reference](https://pkg.go.dev/badge/github.com/jharshman/lifespan.svg)](https://pkg.go.dev/github.com/jharshman/lifespan)
[![CI](https://github.com/jharshman/lifespan/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/jharshman/lifespan/actions/workflows/ci.yaml)

## TL;DR

Package lifespan provides an opinionated method for managing the lifecycle, observability, and coordination of concurrent tasks.
If you find yourself needing to manage the lifecycle of multiple goroutines and are writing and re-writing "glue" code 
to keep them organized, this package might be for you.

> "Never start a goroutine without knowing how it will stop"
    - Dave Cheney

## Quick Start

The package builds on the concept that every goroutine has a "LifeSpan". The LifeSpan is a simple data type that holds channels
of communication to its associated goroutine. By way of the LifeSpan, the user can aggregate errors, logs, and signal goroutines to 
gracefully terminate and also be informed when a goroutine ends.

### The LifeSpan
```golang
// LifeSpan holds the communication channels and context for a runnable task.
type LifeSpan struct {
	// Sig and Ack are the primary control channels. 
	// For example, you can write to Sig to signal to close, and read from Ack to acknowledge.
	Sig, Ack chan struct{}
	// ErrBus is a unique channel that a runnable task can write to.
	// All messages written here are aggregated to the CentralErrorBus.
	ErrBus chan Error
	// Default logger with extra context injected via Run.
	Logger *slog.Logger
}
```

Running a goroutine and receiving a LifeSpan is as simple as invoking the Run function.

```golang
package main

import (
	"context"
	"time"
	
	"github.com/jharshman/lifespan"
)

func main() {
    span, err := lifespan.Run(context.Background(), func(ctx context.Context, span *lifespan.LifeSpan) { 
		// do work.
		// log and report errors
		// you have access to the span and context here as well.
	})
	
	if err != nil {
		// handle error
	}
	
	// act on span
	// in this case just waiting for 3 seconds and then closing.
	<-time.After(time.Second * 3)
	span.Close()
}
```

### Groups

You can logically group goroutines together using `NewGroup`. When running as part of a group,
all goroutines have access to the group's `group_id` which is available when logging or reporting errors.
All LifeSpans within a group are addressable through the Group via their `job_id`.

```golang
package main

import (
	"context"
	
	"github.com/jharshman/lifespan"
)

func main() {
	
	// let us say that job1 to 5 are defined.
	// to group and run them, we can do the following
	group := lifespan.NewGroup(job1, job2, job3, job4, job5)
	
	// Run the group
	group.Start()
}
```

### Error Aggregation

One challenge LifeSpan tries to solve is error reporting and aggregation from multiple
sources. When you have a number of goroutines running, how do you report and collect errors? 
How do you action them? This package provides a `CentralMessageBus` for errors.
Errors emitted by individual goroutines are aggregated into this Message Bus which can be read from
by simply subscribing to it. Errors written here are of type `lifespan.Error` which carries
with it important contextual like `job_id` and `group_id` if applicable.

A utility function is provided to make writing errors as simple as possible.
```golang
lifespan.Error(ctx, err)
```

### Logging

Logging is another challenge. Instead of providing a Message Bus implementation for logs,
or an implementation of `slog.Handler`, the default `slog.Logger` is used.
Each LifeSpan receives a clone of the default logger with additional attributes like `job_id` and `group_id` added.
This means if you log with the provided logger on the LifeSpan, you'll always have access to these attributes in the logs.

```golang
span.Logger.Info("some info")
span.Logger.Warn("some warning")
span.Logger.Error("some error")
```

## Contributing

If you're so inclined, Pull Requests are always welcome.

