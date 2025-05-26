[![Go Report Card](https://goreportcard.com/badge/github.com/jharshman/lifespan)](https://goreportcard.com/report/github.com/jharshman/lifespan)
[![Go Reference](https://pkg.go.dev/badge/github.com/jharshman/async.svg)](https://pkg.go.dev/github.com/jharshman/lifespan)

Package lifespan provides the types, functions, and methods to facilitate the safe running
and closing of jobs.

```
myJob := lifespan.Job{
	Run: func() {
		// do my thing
	},
	Close: func() {
		// close my thing
	},
}
```

Running an HTTP server with lifespan.Job might look like the following:

```
myJob := lifespan.Job{
	Run: func() error {
		return s.ListenAndServe()
	},
	Close: func() error {
		return s.Shutdown(context.Background())
	},
}

myJob.RunWithClose()
```

By default, the function defined for lifespan.Job.Close will trigger when a syscall.SIGINT or
syscall.SIGTERM is received. You can modify these defaults by setting your own on the lifespan.Job.

```
myJob := lifespan.Job{
	Run: func() error {
		return s.ListenAndServe()
	},
	Close: func() error {
		return s.Shutdown(context.Background())
	},
	Signals: []os.Signal{syscall.SIGHUP},
}

myJob.RunWithClose()
```

