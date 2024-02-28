package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/naive"
	"github.com/wcharczuk/go-incr/testutil"
)

// testContext returns a test context.
func testContext() context.Context {
	ctx := context.Background()
	ctx = testutil.WithBlueDye(ctx)
	if os.Getenv("INCR_DEBUG_TRACING") != "" {
		ctx = incr.WithTracing(ctx)
	}
	return ctx
}

func testCase(label string, action func()) {
	fmt.Println("---" + label)
	debug.SetTraceback("all")

	done := make(chan struct{})
	go func() {
		defer func() { close(done) }()
		action()
	}()

	select {
	case <-done:
		return
	case <-time.After(15 * time.Second):
		panic("timeout")
	}
}

func noError(err error) {
	if err != nil {
		panic(fmt.Errorf("expected error to be unset: %v", err))
	}
}

func main() {
	testCase("naiive > month_of_values = if burn > 0: cash_balance / burn else 0. Calculate months of values then burn", func() {
		var cacheMu sync.Mutex
		cache := make(map[string]naive.Node[*int])
		cacheGet := func(key string) (naive.Node[*int], bool) {
			cacheMu.Lock()
			defer cacheMu.Unlock()
			v, ok := cache[key]
			return v, ok
		}
		cachePut := func(key string, value naive.Node[*int]) {
			cacheMu.Lock()
			defer cacheMu.Unlock()
			cache[key] = value
		}

		fakeFormula := naive.Var("fakeformula")
		var f func(int) naive.Node[*int]
		f = func(t int) naive.Node[*int] {
			key := fmt.Sprintf("f-%d", t)
			if cached, ok := cacheGet(key); ok {
				return cached
			}

			r := naive.Bind(fakeFormula, func(formula string) naive.Node[*int] {
				if t <= 0 {
					out := 0
					r := naive.Var(&out)
					return r
				}
				bindOutput := naive.Map(func(values ...*int) *int {
					r := values[0]
					if r == nil {
						return nil
					}
					out := *r + 1
					return &out
				}, f(t-1))
				return bindOutput
			})
			cachePut(key, r)
			return r
		}

		// burn(t) = f(t)
		burn := func(t int) naive.Node[*int] {
			return naive.Bind(fakeFormula, func(formula string) naive.Node[*int] {
				return f(t)
			})
		}

		// cashbalance = cashbalance(t-1) - burn(t)
		var cashBalance func(int) naive.Node[*int]
		cashBalance = func(t int) naive.Node[*int] {
			o := naive.Bind(fakeFormula, func(formula string) naive.Node[*int] {
				if t <= 0 {
					out := 0
					r := naive.Var(&out)
					return r
				}
				return naive.Map(func(values ...*int) *int {
					c := values[0]
					b := values[1]
					if c == nil || b == nil {
						return nil
					}
					out := *c - *b
					return &out
				}, cashBalance(t-1), burn(t))
			})
			return o
		}

		monthsOfRunway := func(t int) naive.Node[*int] {
			o := naive.Bind(fakeFormula, func(formula string) naive.Node[*int] {
				zero := 0
				predicateIncr := naive.Map(func(values ...*int) bool {
					val := values[0]
					cmp := values[1]
					if val == nil || cmp == nil {
						return false
					}
					return *val > *cmp
				}, burn(t), naive.Var(&zero))
				return naive.Bind(predicateIncr, func(predicate bool) naive.Node[*int] {
					var out int = 0
					if predicate {
						return naive.Map(func(values ...*int) *int {
							c := values[0]
							b := values[1]
							if c == nil || b == nil {
								return nil
							}
							out = *c / *b
							return &out
						}, cashBalance(t), burn(t))
					}
					return naive.Var(&out)
				})
			})
			return o
		}

		num := 48

		fmt.Println("Calculating months of values for t= 1 to 48")
		start := time.Now()
		for i := 0; i < num; i++ {
			o := monthsOfRunway(i)
			_ = o.Value()
		}

		elapsed := time.Since(start)
		fmt.Printf("Calculating months of values took %s \n", elapsed)

		fmt.Println("Calculating burn for t= 1 to 48")
		start = time.Now()
		for i := 0; i < num; i++ {
			o := burn(i)
			_ = o.Value()
		}
		elapsed = time.Since(start)
		fmt.Printf("Calculating burn took %s \n", elapsed)
	})

	testCase("naiive > month_of_values = if burn > 0: cash_balance / burn else 0. Calculate months of values then burn", func() {
		var cacheMu sync.Mutex
		cache := make(map[string]naive.Node[*int])
		cacheGet := func(key string) (naive.Node[*int], bool) {
			cacheMu.Lock()
			defer cacheMu.Unlock()
			v, ok := cache[key]
			return v, ok
		}
		cachePut := func(key string, value naive.Node[*int]) {
			cacheMu.Lock()
			defer cacheMu.Unlock()
			cache[key] = value
		}

		fakeFormula := naive.Var("fakeformula")
		var f func(int) naive.Node[*int]
		f = func(t int) naive.Node[*int] {
			key := fmt.Sprintf("f-%d", t)
			if cached, ok := cacheGet(key); ok {
				return cached
			}

			r := naive.Bind(fakeFormula, func(formula string) naive.Node[*int] {
				if t <= 0 {
					out := 0
					r := naive.Var(&out)
					return r
				}
				bindOutput := naive.Map(func(values ...*int) *int {
					r := values[0]
					if r == nil {
						return nil
					}
					out := *r + 1
					return &out
				}, f(t-1))
				return bindOutput
			})
			cachePut(key, r)
			return r
		}

		// burn(t) = f(t)
		burn := func(t int) naive.Node[*int] {
			return naive.Bind(fakeFormula, func(formula string) naive.Node[*int] {
				return f(t)
			})
		}

		// cashbalance = cashbalance(t-1) - burn(t)
		var cashBalance func(int) naive.Node[*int]
		cashBalance = func(t int) naive.Node[*int] {
			o := naive.Bind(fakeFormula, func(formula string) naive.Node[*int] {
				if t <= 0 {
					out := 0
					r := naive.Var(&out)
					return r
				}
				return naive.Map(func(values ...*int) *int {
					c := values[0]
					b := values[1]
					if c == nil || b == nil {
						return nil
					}
					out := *c - *b
					return &out
				}, cashBalance(t-1), burn(t))
			})
			return o
		}

		monthsOfRunway := func(t int) naive.Node[*int] {
			o := naive.Bind(fakeFormula, func(formula string) naive.Node[*int] {
				zero := 0
				predicateIncr := naive.Map(func(values ...*int) bool {
					val := values[0]
					cmp := values[1]
					if val == nil || cmp == nil {
						return false
					}
					return *val > *cmp
				}, burn(t), naive.Var(&zero))
				return naive.Bind(predicateIncr, func(predicate bool) naive.Node[*int] {
					var out int = 0
					if predicate {
						return naive.Map(func(values ...*int) *int {
							c := values[0]
							b := values[1]
							if c == nil || b == nil {
								return nil
							}
							out = *c / *b
							return &out
						}, cashBalance(t), burn(t))
					}
					return naive.Var(&out)
				})
			})
			return o
		}

		num := 48

		fmt.Println("Calculating burn for t= 1 to 48")
		start := time.Now()
		for i := 0; i < num; i++ {
			o := burn(i)
			_ = o.Value()
		}

		elapsed := time.Since(start)
		fmt.Printf("Calculating burn took %s \n", elapsed)

		fmt.Println("Calculating months of values for t= 1 to 48")
		start = time.Now()
		for i := 0; i < num; i++ {
			o := monthsOfRunway(i)
			_ = o.Value()
		}
		elapsed = time.Since(start)
		fmt.Printf("Calculating months of values took %s \n", elapsed)
	})

	testCase("month_of_values = if burn > 0: cash_balance / burn else 0. Calculate months of values then burn", func() {
		ctx := testContext()
		graph := incr.New()

		var cacheMu sync.Mutex
		cache := make(map[string]incr.Incr[*int])
		cacheGet := func(key string) (incr.Incr[*int], bool) {
			cacheMu.Lock()
			defer cacheMu.Unlock()
			v, ok := cache[key]
			return v, ok
		}
		cachePut := func(key string, value incr.Incr[*int]) {
			cacheMu.Lock()
			defer cacheMu.Unlock()
			cache[key] = value
		}

		fakeFormula := incr.Var(graph, "fakeformula")
		fakeFormula.Node().SetLabel("fakeformula")
		var f func(incr.Scope, int) incr.Incr[*int]
		f = func(bs incr.Scope, t int) incr.Incr[*int] {
			key := fmt.Sprintf("f-%d", t)
			if cached, ok := cacheGet(key); ok {
				return cached
			}

			r := incr.Bind(bs, fakeFormula, func(bs incr.Scope, formula string) incr.Incr[*int] {
				if t <= 0 {
					out := 0
					r := incr.Return(bs, &out)
					r.Node().SetLabel("f-0")
					return r
				}
				bindOutput := incr.Map(bs, f(bs, t-1), func(r *int) *int {
					if r == nil {
						return nil
					}
					out := *r + 1
					return &out
				})
				bindOutput.Node().SetLabel(fmt.Sprintf("map-f-%d", t))
				return bindOutput
			})
			r.Node().SetLabel(fmt.Sprintf("f(%d)", t))
			cachePut(key, r)
			return r
		}

		// burn(t) = f(t)
		burn := func(bs incr.Scope, t int) incr.Incr[*int] {
			return incr.Bind(bs, fakeFormula, func(bs incr.Scope, formula string) incr.Incr[*int] {
				return f(bs, t)
			})
		}

		// cashbalance = cashbalance(t-1) - burn(t)
		var cashBalance func(bs incr.Scope, t int) incr.Incr[*int]
		cashBalance = func(bs incr.Scope, t int) incr.Incr[*int] {
			o := incr.Bind(bs, fakeFormula, func(bs incr.Scope, formula string) incr.Incr[*int] {
				if t <= 0 {
					out := 0
					r := incr.Return(bs, &out)
					return r
				}
				return incr.Map2(bs, cashBalance(bs, t-1), burn(bs, t), func(c *int, b *int) *int {
					if c == nil || b == nil {
						return nil
					}
					out := *c - *b
					return &out
				})
			})
			o.Node().SetLabel(fmt.Sprintf("cash_balance(%d)", t))
			return o
		}

		monthsOfRunway := func(bs incr.Scope, t int) incr.Incr[*int] {
			o := incr.Bind(bs, fakeFormula, func(bs incr.Scope, formula string) incr.Incr[*int] {
				zero := 0
				predicateIncr := incr.Map2(bs, burn(bs, t), incr.Return(bs, &zero), func(val *int, cmp *int) bool {
					if val == nil || cmp == nil {
						return false
					}
					return *val > *cmp
				})
				return incr.Bind(bs, predicateIncr, func(bs incr.Scope, predicate bool) incr.Incr[*int] {
					var out int = 0
					if predicate {
						return incr.Map2(bs, cashBalance(bs, t), burn(bs, t), func(c *int, b *int) *int {
							if c == nil || b == nil {
								return nil
							}
							out = *c / *b
							return &out
						})
					}
					return incr.Return(bs, &out)
				})
			})

			o.Node().SetLabel("months_of_values")
			return o
		}
		num := 48

		fmt.Println("Calculating months of values for t= 1 to 48")
		start := time.Now()
		for i := 0; i < num; i++ {
			o := monthsOfRunway(graph, i)
			obs := incr.MustObserve(graph, o)
			obs.Node().SetLabel(fmt.Sprintf("observer(%d)", i))
		}

		err := graph.Stabilize(ctx)
		noError(err)
		elapsed := time.Since(start)
		fmt.Printf("Calculating months of values took %s \n", elapsed)

		fmt.Println("Calculating burn for t= 1 to 48")
		start = time.Now()
		for i := 0; i < num; i++ {
			o := burn(graph, i)
			obs := incr.MustObserve(graph, o)
			obs.Node().SetLabel(fmt.Sprintf("observer(%d)", i))
		}

		err = graph.Stabilize(ctx)
		noError(err)
		elapsed = time.Since(start)
		fmt.Printf("Calculating burn took %s \n", elapsed)
	})

	testCase("month_of_values = if burn > 0: cash_balance / burn else 0. Calculate burn then months of values", func() {
		ctx := testContext()
		graph := incr.New(
			incr.OptGraphMaxHeight(1024),
		)

		var cacheMu sync.Mutex
		cache := make(map[string]incr.Incr[*int])
		cacheGet := func(key string) (incr.Incr[*int], bool) {
			cacheMu.Lock()
			defer cacheMu.Unlock()
			v, ok := cache[key]
			return v, ok
		}
		cachePut := func(key string, value incr.Incr[*int]) {
			cacheMu.Lock()
			defer cacheMu.Unlock()
			cache[key] = value
		}

		fakeFormula := incr.Var(graph, "fakeformula")
		fakeFormula.Node().SetLabel("fakeformula")
		var f func(incr.Scope, int) incr.Incr[*int]
		f = func(bs incr.Scope, t int) incr.Incr[*int] {
			key := fmt.Sprintf("f-%d", t)
			if cached, ok := cacheGet(key); ok {
				return cached
			}
			r := incr.Bind(bs, fakeFormula, func(bs incr.Scope, formula string) incr.Incr[*int] {
				if t <= 0 {
					out := 0
					r := incr.Return(bs, &out)
					r.Node().SetLabel("f-0")
					return r
				}
				bindOutput := incr.Map(bs, f(bs, t-1), func(r *int) *int {
					if r == nil {
						return nil
					}
					out := *r + 1
					return &out
				})
				bindOutput.Node().SetLabel(fmt.Sprintf("map-f-%d", t))
				return bindOutput
			})
			r.Node().SetLabel(fmt.Sprintf("f(%d)", t))
			cachePut(key, r)
			return r
		}

		// burn(t) = f(t)
		burn := func(bs incr.Scope, t int) incr.Incr[*int] {
			return incr.Bind(bs, fakeFormula, func(bs incr.Scope, formula string) incr.Incr[*int] {
				return f(bs, t)
			})
		}

		// cashbalance = cashbalance(t-1) - burn(t)
		var cashBalance func(bs incr.Scope, t int) incr.Incr[*int]
		cashBalance = func(bs incr.Scope, t int) incr.Incr[*int] {
			o := incr.Bind(bs, fakeFormula, func(bs incr.Scope, formula string) incr.Incr[*int] {
				if t <= 0 {
					out := 0
					r := incr.Return(bs, &out)
					return r
				}
				return incr.Map2(bs, cashBalance(bs, t-1), burn(bs, t), func(c *int, b *int) *int {
					if c == nil || b == nil {
						return nil
					}
					out := *c - *b
					return &out
				})
			})
			o.Node().SetLabel(fmt.Sprintf("cash_balance(%d)", t))
			return o
		}

		// monthofvalues = if burn > 0 then cashbalance / burn else 0
		monthsOfRunway := func(bs incr.Scope, t int) incr.Incr[*int] {
			o := incr.Bind(bs, fakeFormula, func(bs incr.Scope, formula string) incr.Incr[*int] {
				zero := 0
				predicateIncr := incr.Map2(bs, burn(bs, t), incr.Return(bs, &zero), func(val *int, cmp *int) bool {
					if val == nil || cmp == nil {
						return false
					}
					return *val > *cmp
				})
				return incr.Bind(bs, predicateIncr, func(bs incr.Scope, predicate bool) incr.Incr[*int] {
					var out int = 0
					if predicate {
						return incr.Map2(bs, cashBalance(bs, t), burn(bs, t), func(c *int, b *int) *int {
							if c == nil || b == nil {
								return nil
							}
							out = *c / *b
							return &out
						})
					}
					return incr.Return(bs, &out)
				})
			})
			o.Node().SetLabel("months_of_values")
			return o
		}
		num := 48

		fmt.Println("Calculating burn for t= 1 to 48")
		start := time.Now()
		for i := 0; i < num; i++ {
			o := burn(graph, i)
			incr.MustObserve(graph, o)
		}

		err := graph.Stabilize(ctx)
		noError(err)
		elapsed := time.Since(start)
		fmt.Printf("Calculating burn took %s\n", elapsed)

		fmt.Println("Calculating months of values for t= 1 to 48")
		start = time.Now()
		for i := 0; i < num; i++ {
			o := monthsOfRunway(graph, i)
			incr.MustObserve(graph, o)
		}

		err = graph.Stabilize(ctx)
		noError(err)
		elapsed = time.Since(start)
		fmt.Printf("Calculating months of values took %s \n", elapsed)
	})

	testCase("node amplification yields slower and slower stabilization", func() {
		ctx := testContext()
		graph := incr.New(
			incr.OptGraphMaxHeight(1024),
		)
		var cacheMu sync.Mutex
		cache := make(map[string]incr.Incr[*int])
		cacheGet := func(key string) (incr.Incr[*int], bool) {
			cacheMu.Lock()
			defer cacheMu.Unlock()
			v, ok := cache[key]
			return v, ok
		}
		cachePut := func(key string, value incr.Incr[*int]) {
			cacheMu.Lock()
			defer cacheMu.Unlock()
			cache[key] = value
		}

		fakeFormula := incr.Var(graph, "fakeformula")
		fakeFormula.Node().SetLabel("fakeformula")
		var f func(incr.Scope, int) incr.Incr[*int]
		f = func(bs incr.Scope, t int) incr.Incr[*int] {
			key := fmt.Sprintf("f-%d", t)
			if cached, ok := cacheGet(key); ok {
				return cached
			}
			r := incr.Bind(bs, fakeFormula, func(bs incr.Scope, formula string) incr.Incr[*int] {
				if t <= 0 {
					out := 0
					r := incr.Return(bs, &out)
					r.Node().SetLabel("f-0")
					return r
				}
				bindOutput := incr.Map(bs, f(bs, t-1), func(r *int) *int {
					if r == nil {
						return nil
					}
					out := *r + 1
					return &out
				})
				bindOutput.Node().SetLabel(fmt.Sprintf("map-f-%d", t))
				return bindOutput
			})
			r.Node().SetLabel(fmt.Sprintf("f(%d)", t))
			cachePut(key, r)
			return r
		}

		// burn(t) = f(t)
		burn := func(bs incr.Scope, t int) incr.Incr[*int] {
			return incr.Bind(bs, fakeFormula, func(bs incr.Scope, formula string) incr.Incr[*int] {
				return f(bs, t)
			})
		}

		// cashbalance = cashbalance(t-1) - burn(t)
		var cashBalance func(bs incr.Scope, t int) incr.Incr[*int]
		cashBalance = func(bs incr.Scope, t int) incr.Incr[*int] {
			o := incr.Bind(bs, fakeFormula, func(bs incr.Scope, formula string) incr.Incr[*int] {
				if t <= 0 {
					out := 0
					r := incr.Return(bs, &out)
					return r
				}
				return incr.Map2(bs, cashBalance(bs, t-1), burn(bs, t), func(c *int, b *int) *int {
					if c == nil || b == nil {
						return nil
					}
					out := *c - *b
					return &out
				})
			})
			o.Node().SetLabel(fmt.Sprintf("cash_balance(%d)", t))
			return o
		}

		// monthofvalues = if burn > 0 then cashbalance / burn else 0
		monthsOfRunway := func(bs incr.Scope, t int) incr.Incr[*int] {
			o := incr.Bind(bs, fakeFormula, func(bs incr.Scope, formula string) incr.Incr[*int] {
				zero := 0
				predicateIncr := incr.Map2(bs, burn(bs, t), incr.Return(bs, &zero), func(val *int, cmp *int) bool {
					if val == nil || cmp == nil {
						return false
					}
					return *val > *cmp
				})
				return incr.Bind(bs, predicateIncr, func(bs incr.Scope, predicate bool) incr.Incr[*int] {
					var out int = 0
					if predicate {
						return incr.Map2(bs, cashBalance(bs, t), burn(bs, t), func(c *int, b *int) *int {
							if c == nil || b == nil {
								return nil
							}
							out = *c / *b
							return &out
						})
					}
					return incr.Return(bs, &out)
				})
			})
			o.Node().SetLabel("months_of_values")
			return o
		}

		w := func(bs incr.Scope, t int) incr.Incr[*int] {
			key := fmt.Sprintf("w-%d", t)
			if cached, ok := cacheGet(key); ok {
				return cached // incr.WithinScope(bs, cache[key])
			}

			r := incr.Bind(bs, incr.Var(bs, "fakeformula"), func(bs incr.Scope, formula string) incr.Incr[*int] {
				out := 1
				return incr.Return(bs, &out)
			})
			r.Node().SetLabel(fmt.Sprintf("w(%d)", t))
			cachePut(key, r)
			return r
		}

		max_t := 50

		// baseline
		start := time.Now()

		for i := 0; i < max_t; i++ {
			o := monthsOfRunway(graph, i)
			_ = incr.MustObserve(graph, o)
		}
		_ = graph.Stabilize(ctx)
		elapsed := time.Since(start)
		fmt.Printf("Baseline calculation of months of values for t= %d to %d took %s\n", 0, max_t, elapsed)

		maxMultiplier := 10
		for k := 1; k <= maxMultiplier; k++ {
			graph := incr.New(
				incr.OptGraphMaxHeight(1024),
			)
			num := 5000 * k

			for i := 0; i < num; i++ {
				o := w(graph, i)
				incr.MustObserve(graph, o)
			}
			start = time.Now()
			for i := 0; i < max_t; i++ {
				o := monthsOfRunway(graph, i)
				_ = incr.MustObserve(graph, o)
			}
			_ = graph.Stabilize(ctx)

			elapsed = time.Since(start)
			fmt.Printf("Calculating months of values for t= %d to %d took %s when prior_count(observed nodes) >%d\n", 0, max_t, elapsed, num)
			fmt.Printf("Graph node count=%d, observer count=%d\n", incr.ExpertGraph(graph).NumNodes(), incr.ExpertGraph(graph).NumObservers())
		}
	})
}
