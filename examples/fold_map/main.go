package main

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil"
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
	ctx := incr.WithTracing(context.Background())

	graph := incr.New()

	data := make(map[incr.Identifier]Order)
	dataInput := incr.Var(data)

	dataInputAdds := incrutil.DiffMapByKeysAdded(dataInput.Read())
	orders := incr.FoldMap(
		dataInputAdds,
		0,
		func(_ incr.Identifier, o Order, v int) int {
			return v + 1
		},
	)
	shares := incr.FoldMap(
		dataInputAdds,
		0,
		func(_ incr.Identifier, o Order, v int) int {
			return v + o.Size
		},
	)
	symbolCounts := incr.FoldMap(
		dataInputAdds,
		make(map[Symbol]int),
		func(_ incr.Identifier, o Order, w map[Symbol]int) map[Symbol]int {
			w[o.Sym]++
			return w
		},
	)

	graph.Observe(dataInput, orders, shares, symbolCounts)
	for x := 0; x < 10; x++ {
		_ = graph.Stabilize(ctx)
		fmt.Println("orders:", orders.Value())
		fmt.Println("shares:", shares.Value())
		fmt.Println("orders by symbol:", symbolCounts.Value())
		fillOrders(data, 2048)
		dataInput.Set(data)
	}
}
