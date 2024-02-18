package main

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/wcharczuk/go-incr"
)

func main() {
	ctx := context.Background()

	values := spigot(ctx)

	g := incr.New()

	currentValue := incr.Var(g, 0.0)
	lastValues := lastN(g, 10, currentValue)
	max := maxOf(g, currentValue, 0.01)
	min := minOf(g, currentValue, 0.01)
	observeMax := incr.MustObserve(g, max)
	observeMax.Node().OnUpdate(func(_ context.Context) {
		fmt.Printf("new max: %0.2f\n", observeMax.Value())
	})
	observeMin := incr.MustObserve(g, min)
	observeMin.Node().OnUpdate(func(_ context.Context) {
		fmt.Printf("new min: %0.2f\n", observeMin.Value())
	})
	critical := anyMatch(g, lastValues, func(v float64) bool {
		return v > 0.99999
	})
	observeLastValues := incr.MustObserve(g, lastValues)
	observeCritical := incr.MustObserve(g, critical)
	observeCritical.Node().OnUpdate(func(_ context.Context) {
		if observeCritical.Value() {
			fmt.Printf("saw critical values!: %v\n", observeLastValues.Value())
		} else {
			fmt.Println("did not see any critical values")
		}
	})
	for {
		value := <-values
		currentValue.Set(value)
		_ = g.Stabilize(ctx)
	}
}

func anyMatch[T any](scope incr.Scope, input incr.Incr[[]T], predicate func(T) bool) incr.Incr[bool] {
	sawAny := incr.Map(scope, input, func(values []T) bool {
		for _, v := range values {
			if predicate(v) {
				return true
			}
		}
		return false
	})
	co := incr.Cutoff(scope, sawAny, func(before, after bool) bool {
		return before == after
	})
	return co
}

type Numbers interface {
	~int | ~float64
}

func maxOf[T Numbers](scope incr.Scope, source incr.Incr[T], epsilon T) incr.Incr[T] {
	var max T
	maxValue := incr.Map(scope, source, func(value T) T {
		if value > max {
			max = value
		}
		return max
	})
	return incr.Cutoff(scope, maxValue, func(prev, next T) bool {
		return absDelta(prev, next) < epsilon
	})
}

func minOf[T Numbers](scope incr.Scope, source incr.Incr[T], epsilon T) incr.Incr[T] {
	var min T
	var didSetMin bool
	minValue := incr.Map(scope, source, func(value T) T {
		if value < min || !didSetMin {
			min = value
			didSetMin = true
		}
		return min
	})
	return incr.Cutoff(scope, minValue, func(prev, next T) bool {
		return absDelta(prev, next) < epsilon
	})
}

func absDelta[T Numbers](a, b T) T {
	if a > b {
		return a - b
	}
	return b - a
}

func lastN[T any](scope incr.Scope, n int, source incr.Incr[T]) incr.Incr[[]T] {
	rolling := make([]T, n)
	var tail, size = 0, 0
	push := func(v T) {
		rolling[tail] = v
		tail = (tail + 1) % n
		if size < n {
			size++
		}
	}
	accum := incr.Map(scope, source, func(cv T) []T {
		push(cv)
		return rolling
	})
	co := incr.Cutoff(scope, accum, func(_, _ []T) bool {
		return size < n
	})
	return co
}

func spigot(ctx context.Context) <-chan float64 {
	output := make(chan float64)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			newValue := rand.Float64()
			select {
			case output <- newValue:
			case <-ctx.Done():
				return
			}
		}
	}()

	return output
}
