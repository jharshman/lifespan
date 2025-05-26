/*
Package lifespan helps to facilitate the safe running and closing of Jobs.

To get started, let's take a look at lifespan.Job. As you can see in the example below
we are defining a lifespan.Job called myJob, and stubbing out the Run and Close fields.
These fields are functions that will control the running and safe closing of your Job.

	myJob := lifespan.Job{
		Run: func() {
			// do my thing
		},
		Close: func() {
			// close my thing
		},
	}

Running an HTTP server with lifespan.Job might look like the following:

	myJob := lifespan.Job{
		Run: func() error {
			return http.ListenAndServe()
		},
		Close: func() error {
			return s.Shutdown(context.Background())
		},
	}

	myJob.RunAndClose()

The provided lifespan.Job implements the SafeCloser interface.
You can use this to make your own implementation suit your needs.
*/
package lifespan

import (
	"os"
	"os/signal"
)

type SafeCloser interface {
	RunWithClose() (err chan error)
}

type Job struct {
	// Run And Close functions.
	Run   func() error
	Close func() error
	// Signals is a slice of os.Signal to notify on.
	Signals []os.Signal
}

func (j *Job) RunWithClose() (err chan error) {
	sig := make(chan int, 1)
	ack := make(chan int, 1)
	err = make(chan error, 1)

	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, j.Signals...)

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

	go func() {
		for {
			select {
			case <-closeChan:
				sig <- 1
			case <-ack:
				break
			}
		}
	}()

	return
}
