package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

func testCase(label string, action func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "panic recovered: %v\n", r)
		}
	}()
	fmt.Println("---" + label)
	action()
}

func skipTestCase(label string, action func()) {
	fmt.Println("--- SKIPPING -" + label)
}

func noError(err error) {
	if err != nil {
		panic(fmt.Errorf("expected error to be unset: %v", err))
	}
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
			// return cache[key]
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

	// burn := func(bs *incr.BindScope, t int) incr.Incr[*int] {
	// 	key := fmt.Sprintf("burn-%d", t)
	// 	if _, ok := cache[key]; ok {
	// 		return incr.WithinBindScope(bs, cache[key])
	// 	}
	// 	o := incr.Bind(bs, fakeFormula, func(bs *incr.BindScope, formula string) incr.Incr[*int] {
	// 		return f(bs, t)
	// 	})
	// 	o.Node().SetLabel(key)
	// 	cache[key] = o
	// 	return o
	// }

	// cashbalance = cashbalance(t-1) - burn(t)
	var cashBalance func(bs *incr.BindScope, t int) incr.Incr[*int]
	cashBalance = func(bs *incr.BindScope, t int) incr.Incr[*int] {
		o := incr.Bind(bs, fakeFormula, func(bs *incr.BindScope, formula string) incr.Incr[*int] {
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

	// monthofrunway = if burn > 0 then cashbalance / burn else 0
	monthsOfRunway := func(bs *incr.BindScope, t int) incr.Incr[*int] {
		o := incr.Bind(bs, fakeFormula, func(bs *incr.BindScope, formula string) incr.Incr[*int] {
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
		o.Node().SetLabel("months_of_runway")
		return o
	}

	testCase("month_of_runway = if burn > 0: cash_balance / burn else 0. Calculate months of runway then burn", func() {
		num := 24
		graph := incr.New()

		fmt.Println("Calculating months of runway for t= 1 to 24")
		start := time.Now()
		for i := 0; i < num; i++ {
			o := monthsOfRunway(incr.Root(), i)
			obs := incr.Observe(incr.Root(), graph, o)
			obs.Node().SetLabel(fmt.Sprintf("observer(%d)", i))
		}

		err := graph.Stabilize(ctx)
		noError(err)
		_ = dumpDot(graph, homedir("integration_00_00.png"))
		elapsed := time.Since(start)
		fmt.Printf(fmt.Sprintf("Calculating months of runway took %s \n", elapsed))

		fmt.Println("Calculating burn for t= 1 to 24")
		start = time.Now()
		for i := 0; i < num; i++ {
			o := burn(incr.Root(), i)
			obs := incr.Observe(incr.Root(), graph, o)
			obs.Node().SetLabel(fmt.Sprintf("observer(%d)", i))
		}

		err = graph.Stabilize(ctx)
		noError(err)
		_ = dumpDot(graph, homedir("integration_00_01.png"))
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
		_ = dumpDot(graph, homedir("integration_01_00.png"))
		elapsed := time.Since(start)
		fmt.Printf(fmt.Sprintf("Calculating burn took %s \n", elapsed))

		fmt.Println("Calculating months of runway for t= 1 to 24")
		start = time.Now()
		for i := 0; i < num; i++ {
			o := monthsOfRunway(incr.Root(), i)
			incr.Observe(incr.Root(), graph, o)
		}

		err = graph.Stabilize(ctx)
		_ = dumpDot(graph, homedir("integration_01_01.png"))
		noError(err)
		elapsed = time.Since(start)
		fmt.Printf(fmt.Sprintf("Calculating months of runway took %s \n", elapsed))
	})
}

func homedir(filename string) string {
	var rootDir string
	if rootDir = os.Getenv("INCR_DEBUG_DOT_ROOT"); rootDir == "" {
		rootDir = os.ExpandEnv("$HOME/Desktop")
	}
	return filepath.Join(rootDir, filename)
}

func dumpDot(g *incr.Graph, path string) error {
	if os.Getenv("INCR_DEBUG_DOT") != "true" {
		return nil
	}

	dotContents := new(bytes.Buffer)
	if err := incr.Dot(dotContents, g); err != nil {
		return err
	}
	dotOutput, err := os.Create(os.ExpandEnv(path))
	if err != nil {
		return err
	}
	defer func() { _ = dotOutput.Close() }()
	dotFullPath, err := exec.LookPath("dot")
	if err != nil {
		return fmt.Errorf("there was an issue finding `dot` in your path; you may need to install the `graphviz` package or similar on your platform: %w", err)
	}

	errOut := new(bytes.Buffer)
	cmd := exec.Command(dotFullPath, "-Tpng")
	cmd.Stdin = dotContents
	cmd.Stdout = dotOutput
	cmd.Stderr = errOut
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%v; %w", errOut.String(), err)
	}
	return nil
}
