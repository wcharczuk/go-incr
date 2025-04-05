package incr

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

type arrayIter[A any] struct {
	values []A
	index  int
}

func (a *arrayIter[A]) Next() (v A, ok bool) {
	if a.index == len(a.values) {
		return
	}

	v = a.values[a.index]
	a.index++
	return v, true
}

func Test_parallelBatch(t *testing.T) {
	var work []string
	for x := 0; x < runtime.NumCPU()<<1; x++ {
		work = append(work, fmt.Sprintf("work-%d", x))
	}

	workIter := &arrayIter[string]{values: work}

	seen := make(map[string]struct{})
	var seenMu sync.Mutex
	err := parallelBatch(testContext(), func(_ context.Context, v string) error {
		seenMu.Lock()
		seen[v] = struct{}{}
		seenMu.Unlock()
		return nil
	}, workIter.Next, runtime.NumCPU())
	testutil.NoError(t, err)
	testutil.Equal(t, len(work), len(seen))

	for x := 0; x < runtime.NumCPU()<<1; x++ {
		key := fmt.Sprintf("work-%d", x)
		_, hasKey := seen[key]
		testutil.Equal(t, true, hasKey)
	}
}

func Test_parallelBatch_error(t *testing.T) {
	var work []string
	for x := 0; x < runtime.NumCPU()<<1; x++ {
		work = append(work, fmt.Sprintf("work-%d", x))
	}
	workIter := &arrayIter[string]{values: work}

	var processed uint32
	err := parallelBatch(testContext(), func(_ context.Context, v string) error {
		atomic.AddUint32(&processed, 1)
		if v == "work-2" {
			return fmt.Errorf("this is only a test")
		}
		return nil
	}, workIter.Next, runtime.NumCPU())
	testutil.Error(t, err)
	testutil.Equal(t, len(work), processed, fmt.Sprintf("work=%d processed=%d", len(work), processed))
}
