package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/wcharczuk/go-incr"
)

// Symbol is a stock symbol identifier.
type Symbol string

// Dir is a order direction
type Dir int

// Dir constants
const (
	Buy Dir = iota
	Sell
)

// Order is an order in the system.
type Order struct {
	ID    incr.Identifier
	Sym   Symbol
	Size  int
	Price float64
	Dir   Dir
}

// Symbols is a list of possible symbols.
var Symbols = []Symbol{
	"GOOG",
	"TWTR",
	"SPY",
	"VIX",
	"MSFT",
	"AAPL",
	"BLND",
}

func randomDir() Dir {
	if rand.Float64() < 0.5 {
		return Buy
	}
	return Sell
}

func randomSize() int {
	return 1 + rand.Intn(255)
}

func randomPrice() float64 {
	return rand.Float64() * 1024.00
}

func randomSymbol() Symbol {
	return Symbols[rand.Intn(len(Symbols)-1)]
}

func fillOrders(output map[incr.Identifier]Order, count int) {
	for x := 0; x < count; x++ {
		orderID := incr.NewIdentifier()
		output[orderID] = Order{
			ID:    orderID,
			Dir:   randomDir(),
			Sym:   randomSymbol(),
			Size:  randomSize(),
			Price: randomPrice(),
		}
	}
}

func main() {
	rand.Seed(time.Now().Unix())

	ctx := incr.WithTracing(context.Background())

	data := make(map[incr.Identifier]Order)
	fillOrders(data, 1024)

	dataInput := incr.Var(data)
	dataInputAdds := MapAdds(dataInput.Read())
	orders := MapFold(
		dataInputAdds,
		0,
		func(_ incr.Identifier, o Order, v int) int {
			return v + 1
		},
	)
	shares := MapFold(
		dataInputAdds,
		0,
		func(_ incr.Identifier, o Order, v int) int {
			return v + o.Size
		},
	)
	symbolCounts := MapFold(
		dataInputAdds,
		make(map[Symbol]int),
		func(_ incr.Identifier, o Order, w map[Symbol]int) map[Symbol]int {
			w[o.Sym]++
			return w
		},
	)

	_ = incr.Stabilize(ctx, shares)
	fmt.Println("orders:", orders.Value())
	fmt.Println("shares:", shares.Value())
	fmt.Println("orders by symbol:", symbolCounts.Value())

	fillOrders(data, 256)
	dataInput.Set(data)

	_ = incr.Stabilize(ctx, shares)
	fmt.Println("orders:", orders.Value())
	fmt.Println("shares:", shares.Value())
	fmt.Println("orders by symbol:", symbolCounts.Value())
}

func MapAdds[K comparable, V any](
	i incr.Incr[map[K]V],
) incr.Incr[map[K]V] {
	o := &mapAddsIncr[K, V]{
		n: incr.NewNode(),
		i: i,
	}
	incr.Link(o, i)
	return o
}

type mapAddsIncr[K comparable, V any] struct {
	n   *incr.Node
	i   incr.Incr[map[K]V]
	val map[K]V
}

func (mfn *mapAddsIncr[K, V]) String() string { return "map_adds[" + mfn.Node().ID().Short() + "]" }

func (mfn *mapAddsIncr[K, V]) Node() *incr.Node { return mfn.n }

func (mfn *mapAddsIncr[K, V]) Value() map[K]V { return mfn.val }

func (mfn *mapAddsIncr[K, V]) Stabilize(_ context.Context) error {
	mfn.val = diffMapAdds(mfn.val, mfn.i.Value())
	return nil
}

// MapFold returns an incremental that takes a map typed incremental as an
// input, an initial value, and a combinator yielding an incremental
// representing the result of the combinator.
//
// Between stabilizations only the _additions_ to the input map will be considered for subsequent folds, and as a result
// just a subset of the computation will be processed each pass.
func MapFold[K comparable, V any, O any](
	i incr.Incr[map[K]V],
	v0 O,
	fn func(K, V, O) O,
) incr.Incr[O] {
	o := &mapFoldIncr[K, V, O]{
		n:   incr.NewNode(),
		i:   i,
		fn:  fn,
		val: v0,
	}
	incr.Link(o, i)
	return o
}

type mapFoldIncr[K comparable, V any, O any] struct {
	n    *incr.Node
	i    incr.Incr[map[K]V]
	fn   func(K, V, O) O
	last map[K]V
	val  O
}

func (mfn *mapFoldIncr[K, V, O]) String() string { return "map_fold[" + mfn.Node().ID().Short() + "]" }

func (mfn *mapFoldIncr[K, V, O]) Node() *incr.Node { return mfn.n }

func (mfn *mapFoldIncr[K, V, O]) Value() O { return mfn.val }

func (mfn *mapFoldIncr[K, V, O]) Stabilize(_ context.Context) error {
	new := mfn.i.Value()
	mfn.val = fold(new, mfn.val, mfn.fn)
	return nil
}

func fold[K comparable, V any, O any](input map[K]V, zero O, fn func(K, V, O) O) (o O) {
	o = zero
	for k, v := range input {
		o = fn(k, v, o)
	}
	return
}

func diffMapAdds[K comparable, V any](m0, m1 map[K]V) (add map[K]V) {
	add = make(map[K]V)
	var ok bool
	if m0 != nil {
		for k, v := range m1 {
			if _, ok = m0[k]; !ok {
				add[k] = v
			}
		}
		return
	}
	for k, v := range m1 {
		add[k] = v
	}
	return
}

func diffMap[K comparable, V any](m0, m1 map[K]V) (add, rem map[K]V) {
	add = make(map[K]V)
	rem = make(map[K]V)
	var ok bool
	if m0 != nil && m1 != nil {
		for k, v := range m1 {
			if _, ok = m0[k]; !ok {
				add[k] = v
			}
		}
		for k, v := range m0 {
			if _, ok = m1[k]; !ok {
				rem[k] = v
			}
		}
		return
	}
	if m0 != nil {
		for k, v := range m0 {
			rem[k] = v
		}
		return
	}
	for k, v := range m1 {
		add[k] = v
	}
	return
}
