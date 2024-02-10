package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/wcharczuk/go-incr"
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

func main() {
	ctx := testContext()
	cache := make(map[string]incr.Incr[*int])

	fakeFormula := incr.Var(incr.Root(), "fakeformula")
	fakeFormula.Node().SetLabel("fakeformula")
	var f func(*incr.BindScope, int) incr.Incr[*int]
	f = func(bs *incr.BindScope, t int) incr.Incr[*int] {
		key := fmt.Sprintf("f-%d", t)
		if _, ok := cache[key]; ok {
			return incr.WithinBindScope(bs, cache[key])
		}

		r := incr.Bind(incr.Root(), fakeFormula, func(bs *incr.BindScope, formula string) incr.Incr[*int] {
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
		cache[key] = r
		return r
	}

	// burn(t) = f(t)
	burn := func(bs *incr.BindScope, t int) incr.Incr[*int] {
		return incr.Bind(bs, fakeFormula, func(bs *incr.BindScope, formula string) incr.Incr[*int] {
			return f(bs, t)
		})
	}

	// cashbalance = cashbalance(t-1) - burn(t)
	var cash_balance func(bs *incr.BindScope, t int) incr.Incr[*int]
	cash_balance = func(bs *incr.BindScope, t int) incr.Incr[*int] {
		return incr.Bind(bs, fakeFormula, func(bs *incr.BindScope, formula string) incr.Incr[*int] {
			if t <= 0 {
				out := 0
				r := incr.Return(bs, &out)
				return r
			}
			return incr.Map2(bs, cash_balance(bs, t-1), burn(bs, t), func(c *int, b *int) *int {
				if c == nil || b == nil {
					return nil
				}
				out := *c - *b
				return &out
			})
		})
	}

	// monthofrunway = if burn > 0 then cashbalance / burn else 0
	months_of_runway := func(bs *incr.BindScope, t int) incr.Incr[*int] {
		return incr.Bind(bs, fakeFormula, func(bs *incr.BindScope, formula string) incr.Incr[*int] {
			zero := 0
			predicateIncr := incr.Map2(bs, burn(bs, t), incr.Return(bs, &zero), func(val *int, cmp *int) bool {
				if val == nil || cmp == nil {
					return false
				}
				return *val > *cmp
			})
			return incr.Bind(bs, predicateIncr, func(bs *incr.BindScope, predicate bool) incr.Incr[*int] {
				var out int = 0
				if predicate {
					return incr.Map2(bs, cash_balance(bs, t), burn(bs, t), func(c *int, b *int) *int {
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
	}

	testCase := func(label string, action func()) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "panic recovered: %v\n", r)
			}
		}()
		fmt.Println("---" + label)
		action()
	}

	noError := func(err error) {
		if err != nil {
			panic(fmt.Errorf("expected error to be unset: %v", err))
		}
	}

	testCase("month_of_runway = if burn > 0: cash_balance / burn else 0. Calculate months of runway then burn", func() {
		num := 24
		graph := incr.New()

		fmt.Println("Calculating months of runway for t= 1 to 24")
		start := time.Now()
		for i := 0; i < num; i++ {
			o := months_of_runway(incr.Root(), i)
			incr.Observe(incr.Root(), graph, o)
		}

		err := graph.Stabilize(ctx)
		noError(err)
		elapsed := time.Since(start)
		fmt.Printf(fmt.Sprintf("Calculating months of runway took %s \n", elapsed))

		fmt.Println("Calculating burn for t= 1 to 24")
		start = time.Now()
		for i := 0; i < num; i++ {
			o := burn(incr.Root(), i)
			incr.Observe(incr.Root(), graph, o)
		}

		err = graph.Stabilize(ctx)
		noError(err)
		elapsed = time.Since(start)
		fmt.Printf(fmt.Sprintf("Calculating burn took %s \n", elapsed))
	})

	testCase("month_of_runway = if burn > 0: cash_balance / burn else 0. Calculate burn then months of runway", func() {
		num := 24
		graph := incr.New()

		fmt.Println("Calculating burn for t= 1 to 24")
		start := time.Now()
		for i := 0; i < num; i++ {
			o := burn(incr.Root(), i)
			incr.Observe(incr.Root(), graph, o)
		}

		err := graph.Stabilize(ctx)
		noError(err)
		elapsed := time.Since(start)
		fmt.Printf(fmt.Sprintf("Calculating burn took %s \n", elapsed))

		fmt.Println("Calculating months of runway for t= 1 to 24")
		start = time.Now()
		for i := 0; i < num; i++ {
			o := months_of_runway(incr.Root(), i)
			incr.Observe(incr.Root(), graph, o)
		}

		err = graph.Stabilize(ctx)
		noError(err)
		elapsed = time.Since(start)
		fmt.Printf(fmt.Sprintf("Calculating months of runway took %s \n", elapsed))
	})
}
