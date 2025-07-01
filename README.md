# Lifespan
[![Go Report Card](https://goreportcard.com/badge/github.com/jharshman/lifespan)](https://goreportcard.com/report/github.com/jharshman/lifespan)
[![Go Reference](https://pkg.go.dev/badge/github.com/jharshman/lifespan.svg)](https://pkg.go.dev/github.com/jharshman/lifespan)
[![CI](https://github.com/jharshman/lifespan/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/jharshman/lifespan/actions/workflows/ci.yaml)

## TL;DR

Package lifespan provides an opinionated method for managing the lifecycle, observability, and coordination of concurrent tasks.

> "Never start a goroutine without knowing how it will stop"
    - Dave Cheney

## Example Usage

In this basic example, we use a closure to define the function that lifespan.Run will execute.
Here we have a simple control loop and the resulting goroutine will print "hello world" and sleep for one second
until it is signaled to close.

```golang
func main() {
	span, _ := lifespan.Run(ctx, func(ctx context.Context, span *lifespan.LifeSpan) {
	LOOP:
		for {
			select {
			case <-ctx.Done():
				break LOOP
			case <-span.Sig:
				break LOOP
			default:
			}
			fmt.Println("hello world")
			time.Sleep(1 * time.Second)
		}
		span.Ack <- struct{}{}
	})

	time.Sleep(5 * time.Second)
	fmt.Println("exiting")
	span.Close()
}
```

Getting a bit more in-depth, we can define custom jobs and even groups of jobs.
Below we define a Job struct and implement the Runnable interface for it. The implementation 
is similar to the previous example except here we are demonstrating that each LifeSpan gets a UUID.

```golang
type Job struct{}

func (j *Job) Run(ctx, span *lifespan.LifeSpan) {
LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case <-span.Sig:
			break LOOP
		default:
		}
		fmt.Printf("hello from Job: %s\n", ctx.Value("job_id").(string))
		time.Sleep(1 * time.Second)
	}
	fmt.Printf("done with Job: %s\n", ctx.Value("job_id").(string))
	span.Ack <- struct{}{}
}
```

Below we take the above implementation and demonstrate different ways to use it.

1. Running a job and responding to an os.Signal like SIGTERM or SIGINT.
2. Running a job and responding to a context timeout.
3. Creating a group of jobs.
4. Stopping select jobs from a group.
5. Stopping remaining jobs in a group.

```golang
func main() {
	
    j1 := &Job{}
	
    // 1. Running a job and responding to an os.Signal like SIGTERM or SIGINT
    
    span, _ := lifespan.Run(context.Background(), j1.Run)
    notify := make(chan os.Signal, 1)
    signal.Notify(notify, syscall.SIGTERM, syscall.SIGINT)
    <-notify
    span.Close()
    
    // 2. Running a job and responding to a context timeout

	ctx, cancel := context.WithTimeout(span.Ctx, 5*time.Second)
    span, _ = lifespan.Run(ctx, j1.Run)
    <-span.Ack
    
    // 3. Creating a group of jobs
    
    j2 := &Job{}
    j3 := &Job{}
    j4 := &Job{}
    j5 := &Job{}
    
    group := lifespan.NewGroup(j1, j2, j3, j4, j5)
    group.Start()
    
    time.Sleep(3 * time.Second)
    
    // 4. Stopping individual jobs within a group
    
    group.Spans[3].Close()
    group.Spans[4].Close()
    
    time.Sleep(3 * time.Second)
    
    // 5. Stop remaining jobs in group
    
    group.Close()
    fmt.Println("all done")
}
```

## Message Aggregation

Lifespan provides methods of Log and Error aggregation via an internal Message Bus.

### Logging

Lifespan provides a Logger on each LifeSpan. This default Logger includes the `job_id` and `group_id` (if applicable) and are
pulled from the context.

```golang
func main() {
    // Pass the logHandler into the Run function
    // Run will put the job_id into the logger so every log can be attributed to the job that emitted it.
    span, _ := lifespan.Run(ctx, func(ctx context.Context, span *lifespan.LifeSpan) {
        // example of log usage
        span.Logger.Info("log at info level") // 2025/06/29 21:27:42 INFO log at info level job_id=8439b094-8192-4d02-a545-887e1bcd0926 group_id=""
        span.Logger.Warn("log at warn level") // 2025/06/29 21:27:42 WARN log at warn level job_id=8439b094-8192-4d02-a545-887e1bcd0926 group_id=""
        span.Logger.Error("log at error level") // 2025/06/29 21:27:42 ERROR log at error level job_id=8439b094-8192-4d02-a545-887e1bcd0926 group_id=""
        })
}
```


### Errors

Lifespan provides a central message bus for errors. Each LifeSpan is given its own channel to write errors into and this is aggregated into
a Message Bus that can be subscribed to.

```golang
    // Creates a logHandler.
    // The lifespan implementation of log/slog.Handler will write to an underlying implementation of MessageBus for Logs.
	
    // Pass the logHandler into the Run function
    // Run will put the job_id into the logger so every log can be attributed to the job that emitted it.
    span, _ := lifespan.Run(ctx, func(ctx context.Context, span *lifespan.LifeSpan) {
		
		span.ErrBus.Publish(ctx, errors.New("some error"))
		
    })
	
	fmt.Println(<-CentralErrorBus.Subscribe())
}
```

## Contributing

If you're so inclined, Pull Requests are always welcome.

