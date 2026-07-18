// Command incremental_map shows a computation over a keyed collection where the cost of
// an update is proportional to what changed rather than to the size of the collection.
//
// The subject is a book of orders keyed by id, with three derived values maintained over
// it: a per-order display line, the total notional, and the largest single order. Orders
// are then added, repriced and filled, and each pass reports how much work it did.
//
// The thing to notice is the recompute counts. A book of 5000 orders with one repriced
// re-runs one per-order transform and adjusts the total in constant time, rather than
// touching 5000 orders to answer either question.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/mapi"
	"github.com/wcharczuk/go-incr/incrutil/pmap"
)

type order struct {
	Symbol   string
	Quantity int
	Price    int // in cents, to keep the arithmetic exact
}

func (o order) notional() int { return o.Quantity * o.Price }

func main() {
	ctx := context.Background()
	g := incr.New()

	// The book has to be a pmap.Map rather than a map[string]order. That is the whole
	// point: a pmap shares structure with the version it was derived from, so the
	// operators below can tell what changed by comparing the two in time proportional
	// to the number of changes. Two builtin maps share nothing, and comparing them
	// means looking at every key however few of them moved.
	book := pmap.New[string, order]()
	for i := range 5000 {
		book = book.Set(fmt.Sprintf("ord-%04d", i), order{
			Symbol:   []string{"AAPL", "MSFT", "GOOG"}[i%3],
			Quantity: 100 + i%50,
			Price:    10_000 + i,
		})
	}
	orders := incr.Var(g, book)

	// Two orders are equal if all their fields are, which is what tells the operators
	// whether a key's value actually moved.
	sameOrder := func(a, b order) bool { return a == b }

	// A per-order transform. Only the entries that changed are recomputed, so this
	// counter is the interesting output.
	var lineRecomputes int
	lines := mapi.MapValues(g, orders, sameOrder, func(id string, o order) string {
		lineRecomputes++
		return fmt.Sprintf("%s %s %d @ %d.%02d", id, o.Symbol, o.Quantity, o.Price/100, o.Price%100)
	})

	// Total notional. Addition has an inverse, so this is maintained by withdrawing the
	// changed order's old contribution and applying its new one: constant time per
	// changed key, whatever the size of the book.
	total := mapi.UnorderedFold(g, orders, 0, sameOrder,
		func(acc int, _ string, o order) int { return acc + o.notional() },
		func(acc int, _ string, o order) int { return acc - o.notional() })

	// Largest single order. A maximum has no inverse -- withdrawing the current largest
	// tells you nothing about the next one -- so this cannot be a running accumulator.
	// mapi.Reduce folds over the collection's own tree instead, which recomputes only
	// the path from the changed entry to the root: O(log n) rather than O(n).
	largest := mapi.Reduce(g, orders, 0,
		func(_ string, o order) int { return o.notional() },
		func(a, b int) int { return max(a, b) })

	// A window over the book in key order, with bounds that are themselves incremental.
	// Moving the window costs the size of the window, not the size of the book.
	window := incr.Var(g, mapi.Bounds[string]{Low: "ord-0000", High: "ord-0004"})
	page := mapi.Subrange(g, lines, window, func(a, b string) bool { return a == b })

	count := mapi.Cardinality(g, orders)

	observedLines := incr.MustObserve(g, lines)
	observedTotal := incr.MustObserve(g, total)
	observedLargest := incr.MustObserve(g, largest)
	observedPage := incr.MustObserve(g, page)
	observedCount := incr.MustObserve(g, count)

	report := func(what string) {
		lineRecomputes = 0
		if err := g.Stabilize(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "%+v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%-28s orders=%-5d total=%-12d largest=%-8d lines recomputed=%d\n",
			what, observedCount.Value(), observedTotal.Value(), observedLargest.Value(), lineRecomputes)
	}

	// The first pass has to build everything, so every order is transformed once.
	report("initial build")

	// Repricing one order recomputes one line, and adjusts the total without revisiting
	// the other 4999.
	book = book.Set("ord-0002", order{Symbol: "AAPL", Quantity: 100, Price: 12_500})
	orders.Set(book)
	report("reprice one order")

	// Adding an order is the same story.
	book = book.Set("ord-9999", order{Symbol: "TSLA", Quantity: 1_000, Price: 25_000})
	orders.Set(book)
	report("add a large order")

	// Filling it removes the largest, which is the case a running accumulator cannot
	// handle: the new maximum has to be found rather than derived.
	book = book.Delete("ord-9999")
	orders.Set(book)
	report("fill the large order")

	// Setting the book to an identical value changes nothing, so nothing recomputes.
	orders.Set(book)
	report("no-op update")

	fmt.Println()
	fmt.Println("first page of the book:")
	for _, line := range sortedValues(observedPage.Value()) {
		fmt.Println(" ", line)
	}

	// Moving the window rebuilds only the new window.
	window.Set(mapi.Bounds[string]{Low: "ord-1000", High: "ord-1002"})
	if err := g.Stabilize(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
	fmt.Println()
	fmt.Println("after moving the window:")
	for _, line := range sortedValues(observedPage.Value()) {
		fmt.Println(" ", line)
	}

	// The whole book is still available where it is genuinely needed; the operators
	// above exist so that answering questions about it does not require reading it.
	fmt.Println()
	fmt.Printf("book holds %d lines\n", observedLines.Value().Len())
}

// sortedValues returns a map's values in key order, which pmap iteration already gives.
func sortedValues(m pmap.Map[string, string]) []string {
	out := make([]string, 0, m.Len())
	for _, value := range m.All() {
		out = append(out, value)
	}
	return out
}
