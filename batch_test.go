package incr

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

func Test_Batch(t *testing.T) {
	work := make(chan string, 10)

	for x := 0; x < 10; x++ {
		work <- fmt.Sprint(x)
	}

	var values []string
	var valuesMu sync.Mutex
	err := Batch(context.Background(), work, func(_ context.Context, v string) error {
		valuesMu.Lock()
		values = append(values, v)
		valuesMu.Unlock()
		return nil
	})
	itsNil(t, err)
	for x := 0; x < 10; x++ {
		itsAny(t, values, fmt.Sprint(x))
	}
}

func Test_Batch_Error(t *testing.T) {
	work := make(chan string, 10)

	for x := 0; x < 10; x++ {
		work <- fmt.Sprint(x)
	}
	err := Batch(context.Background(), work, func(_ context.Context, v string) error {
		if v == "5" {
			return fmt.Errorf("this is just a test")
		}
		return nil
	})
	itsNotNil(t, err)
}

func Test_Batch_Cancel(t *testing.T) {
	work := make(chan string, 10)

	for x := 0; x < 10; x++ {
		work <- fmt.Sprint(x)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		_ = Batch(ctx, work, func(ctx context.Context, _ string) error {
			select {
			case <-ctx.Done():
				return nil
			}
		})
	}()
	cancel()
}
