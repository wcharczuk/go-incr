package incr

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_parallelBatch(t *testing.T) {
	var work []string
	for x := 0; x < runtime.NumCPU()<<1; x++ {
		work = append(work, fmt.Sprintf("work-%d", x))
	}

	var seen []string
	var seenMu sync.Mutex
	err := parallelBatch[string](testContext(), func(_ context.Context, v string) error {
		seenMu.Lock()
		seen = append(seen, v)
		seenMu.Unlock()
		return nil
	}, work...)
	testutil.NoError(t, err)
	testutil.Equal(t, len(work), len(seen))
}

func Test_parallelBatch_error(t *testing.T) {
	var work []string
	for x := 0; x < runtime.NumCPU()<<1; x++ {
		work = append(work, fmt.Sprintf("work-%d", x))
	}

	var processed int
	err := parallelBatch[string](testContext(), func(_ context.Context, v string) error {
		processed++
		if v == "work-2" {
			return fmt.Errorf("this is only a test")
		}
		return nil
	}, work...)
	testutil.Error(t, err)
	testutil.Equal(t, len(work), processed)
}
