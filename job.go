/*
Package lifespan helps to facilitate the safe running and closing of Jobs.

To get started, let's take a look at lifespan.Job. As you can see in the example below
we are defining a lifespan.Job called myJob, and stubbing out the Run and Close fields.
These fields are functions that will control the running and safe closing of your Job.

	myJob := lifespan.Job{
		Run: func(ctx context.Context) error {
			// do my thing
		},
		Close: func(ctx context.Context) error {
			// close my thing
		},
	}

Running an HTTP server with lifespan.Job might look like the following:

	myJob := lifespan.Job{
		Run: func(ctx context.Context) error {
			return http.ListenAndServe()
		},
		Close: func(ctx.Context.Context) error {
			return s.Shutdown(context.Background())
		},
	}

	myJob.RunAndClose()

The provided lifespan.Job implements the SafeCloser interface.
You can use this to make your own implementation suit your needs.
*/
package lifespan

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// SafeCloser defines the behavior expected for a Job or other SafeCloser implementation.
// It is expected that RunWithClose not only runs a thing, but also knows how to safely exit it.
type SafeCloser interface {
	RunWithClose(ctx context.Context) (err chan error)
}

// Job is a simple implementation of SafeCloser. The Run and Close methods define what a task should do and how it
// should end. Job also contains an array of os.Signal. These are used by RunWithClose to determine when to trigger the Close()
// function.
type Job struct {
	// Run And Close functions.
	Run   func() error
	Close func() error
	// Signals is a slice of os.Signal to notify on.
	Signals []os.Signal
}

// RunWithClose is the implementation of SafeCloser for Job. In this method, the Job.Run function is called within a
// goroutine. When a signal is received indicating it is time for Job.Run to end, the Job.Close function is invoked.
// An acknowledgement indicating the Job.Close method is finished is then sent to the control loop.
// This method is non-blocking. If any waiting or blocking need occur, it must happen outside of this implementation.
func (j *Job) RunWithClose(ctx context.Context) (err chan error) {
	sig := make(chan int, 1)
	ack := make(chan int, 1)
	err = make(chan error, 1)

	closeChan := make(chan os.Signal, 1)
	if len(j.Signals) == 0 {
		j.Signals = []os.Signal{
			syscall.SIGINT,
			syscall.SIGTERM,
		}
	}
	signal.Notify(closeChan, j.Signals...)

	// Run and Close
	go func() {
		go func() {
			if e := j.Run(); e != nil {
				err <- e
			}
		}()
		<-sig
		if e := j.Close(); e != nil {
			err <- e
		}
		ack <- 1
	}()

	// Control
	go func() {
	LOOP:
		for {
			select {
			case <-closeChan:
				sig <- 1
			case <-ack:
				break LOOP
			case <-ctx.Done():
				break LOOP
			}
		}
	}()

	return
}
