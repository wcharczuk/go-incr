package main

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/mapi"
)

// Symbol is a stock symbol identifier.
type Symbol string

// Dir is a order direction.
type Dir int

// Dir constants.
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
	"NVDA",
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
	ctx := context.Background()
	graph := incr.New()

	data := make(map[incr.Identifier]Order)
	dataInput := incr.Var(graph, data)

	dataInputAdds := mapi.Added(graph, dataInput)

	orders := incr.Map(
		graph,
		dataInputAdds,
		func(added map[incr.Identifier]Order) (total int) {
			total += len(added)
			return
		},
	)

	shares := incr.Map(
		graph,
		dataInputAdds,
		func(added map[incr.Identifier]Order) (total int) {
			for _, o := range added {
				total += o.Size
			}
			return
		},
	)

	symbolCounts := incr.Map(
		graph,
		dataInputAdds,
		func(added map[incr.Identifier]Order) (output map[Symbol]int) {
			output = make(map[Symbol]int)
			for _, o := range added {
				output[o.Sym]++
			}
			return output
		},
	)

	ordersObs := incr.MustObserve(graph, orders)
	sharesObs := incr.MustObserve(graph, shares)
	symbolObs := incr.MustObserve(graph, symbolCounts)

	for x := 0; x < 10; x++ {
		_ = graph.Stabilize(ctx)
		fmt.Println("orders:", ordersObs.Value())
		fmt.Println("shares:", sharesObs.Value())
		fmt.Println("orders by symbol:", symbolObs.Value())
		fillOrders(data, 2048)
		dataInput.Set(data)
	}
}
