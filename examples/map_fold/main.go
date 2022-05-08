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

	data := make(map[incr.Identifier]Order)
	fillOrders(data, 1024)

	dataInput := incr.Var(data)

	shares := MapFold(dataInput.Read(), func(_ incr.Identifier, o Order, v int) int {
		return v + o.Size
	})

	_ = incr.Stabilize(context.Background(), shares)
	fmt.Println(shares.Value())

	fillOrders(data, 256)
	dataInput.Set(data)

	_ = incr.Stabilize(context.Background(), shares)
	fmt.Println(shares.Value())
}

func MapFold[K comparable, V any, O any](i incr.Incr[map[K]V], fn func(K, V, O) O) incr.Incr[O] {
	o := &mapFoldIncr[K, V, O]{
		n:  incr.NewNode(),
		i:  i,
		fn: fn,
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

func (mfn *mapFoldIncr[K, V, O]) Node() *incr.Node { return mfn.n }

func (mfn *mapFoldIncr[K, V, O]) Value() O { return mfn.val }

func (mfn *mapFoldIncr[K, V, O]) Stabilize(_ context.Context) error {
	new := mfn.i.Value()
	adds, _ := diffMap(mfn.last, new)
	mfn.val = fold(adds, mfn.val, mfn.fn)
	return nil
}

func fold[K comparable, V any, O any](input map[K]V, zero O, fn func(K, V, O) O) (o O) {
	o = zero
	for k, v := range input {
		o = fn(k, v, o)
	}
	return
}

// diffMap returns the additions and removals for two different "versions" of a given map.
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
