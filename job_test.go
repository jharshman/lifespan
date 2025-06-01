package lifespan_test

import (
  "fmt"
	"testing"
  "time"

	"github.com/jharshman/lifespan"
)

func Test_Run(t *testing.T) {
  span := lifespan.Run(func(span *lifespan.LifeSpan) {
    LOOP: for {
      select {
      case <-span.Ctx.Done():
        break LOOP
      case <-span.Sig:
        break LOOP
      default:
      }
      fmt.Println("hello from Run function")
      time.Sleep(1 * time.Second)
    }
    span.Ack <- struct{}{}
  })

  time.Sleep(5 * time.Second)
  fmt.Println("exiting")
  span.Close()

}

