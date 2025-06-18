# Lifespan
[![Go Report Card](https://goreportcard.com/badge/github.com/jharshman/lifespan)](https://goreportcard.com/report/github.com/jharshman/lifespan)
[![Go Reference](https://pkg.go.dev/badge/github.com/jharshman/lifespan.svg)](https://pkg.go.dev/github.com/jharshman/lifespan)

## TL;DR

Package lifespan provides an opinionated (yet hopefully flexible enough) method for defining and running synchronous or asynchronous tasks.
The goal is to encourage good patterns when using goroutines.

> "Never start a goroutine without knowing how it will stop"
    - Dave Cheney

## Example Usage

In this basic example, we use a closure to define the function that lifespan.Run will execute.
Here we have a simple control loop and the resulting goroutine will print "hello world" and sleep for one second
until it is signaled to close.

```golang
func main() {
	span := lifespan.Run(func(span *lifespan.LifeSpan) {
	LOOP:
		for {
			select {
			case <-span.Ctx.Done():
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

func (j *Job) Run(span *lifespan.LifeSpan) {
LOOP:
	for {
		select {
		case <-span.Ctx.Done():
			break LOOP
		case <-span.Sig:
			break LOOP
		default:
		}
		fmt.Printf("hello from Job: %s\n", span.UUID.String())
		time.Sleep(1 * time.Second)
	}
	fmt.Printf("done with Job: %s\n", span.UUID.String())
	span.Ack <- struct{}{}
}
```

Here we demonstrate different ways to use the custom Job we defined.

1. Running a job and responding to an os.Signal like SIGTERM or SIGINT.
2. Running a job and responding to a context timeout.
3. Creating a group of jobs.
4. Stopping select jobs from a group.
5. Stopping remaining jobs in a group.

```golang
func main() {

	j1 := &Job{}

    // 1. Running a job and responding to an os.Signal like SIGTERM or SIGINT

	span := lifespan.Run(j1.Run)
	notify := make(chan os.Signal, 1)
	signal.Notify(notify, syscall.SIGTERM, syscall.SIGINT)
	<-notify
	span.Close()

    // 2. Running a job and responding to a context timeout

	span = lifespan.Run(j1.Run)
	// lifespans have contexts and cancel functions. Here we overwrite them with a timeout.
    // We wait for the timeout which will send an Ack once the goroutine has finished.
	span.Ctx, span.Cancel = context.WithTimeout(span.Ctx, 5*time.Second)
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

## Contributing

If you're so inclined, Pull Requests are always welcome.

