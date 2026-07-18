// Command subscriptions is a dashboard whose panels each hold an expensive resource: a
// live subscription to a data source. Panels are added and removed as the user
// reconfigures the view, and the subscriptions have to be opened and closed to match.
//
// This is the case node lifecycle handlers exist for. A node leaving the computation is
// invisible through values: nothing changes, the node simply stops being asked. Without
// [incr.Node.OnBecameUnnecessary] there is no moment at which to close a connection, and
// the usual result is a leak that only shows up as a slowly growing connection count in
// production.
//
// The output tracks opens and closes against the live subscription count, so a leak or a
// double close would be visible rather than implied -- and the feed panics on either, so
// getting the placement of the lifecycle handlers wrong fails loudly. See the comment on
// where the panel nodes are built.
package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/wcharczuk/go-incr"
)

// feed stands in for something that costs money to hold open: a websocket, a database
// cursor, a market data entitlement.
type feed struct {
	name   string
	open   bool
	opens  int
	closes int
}

func (f *feed) Open() {
	if f.open {
		panic("feed " + f.name + " opened twice")
	}
	f.open = true
	f.opens++
}

func (f *feed) Close() {
	if !f.open {
		panic("feed " + f.name + " closed while not open")
	}
	f.open = false
	f.closes++
}

func main() {
	ctx := context.Background()
	g := incr.New(incr.OptGraphMaxHeight(64))

	// The available sources, and the resource each one holds.
	feeds := map[string]*feed{}
	for _, name := range []string{"prices", "trades", "news", "risk", "positions"} {
		feeds[name] = &feed{name: name}
	}

	// A panel node per source, built once, outside any bind.
	//
	// This placement is the whole point. Lifecycle handlers fire for a *node*, and a
	// bind creates fresh nodes every time it rebuilds -- so building the panels inside
	// the bind would give each source a new node per layout change, and the new node
	// becomes necessary before the old one becomes unnecessary. The result is an Open
	// before the matching Close: for a subscription, a duplicate connection. Built once
	// out here, each source has one node whose necessity follows the layout exactly.
	panels := make(map[string]incr.Incr[string], len(feeds))
	for name, source := range feeds {
		panel := incr.Map(g, incr.Return(g, name), func(n string) string {
			return "[" + n + "]"
		})
		panel.Node().OnBecameNecessary(source.Open)
		panel.Node().OnBecameUnnecessary(source.Close)
		panels[name] = panel
	}

	// Which panels the dashboard is showing. Changing this changes which panels are
	// referenced, so it drives a bind.
	layout := incr.Var(g, []string{"prices", "trades"})

	// The bind selects among the panels rather than creating them. A panel that stops
	// being referenced stops being necessary, which is what closes its subscription; no
	// code here diffs the old layout against the new one.
	dashboard := incr.Bind(g, layout, func(bs incr.Scope, names []string) incr.Incr[string] {
		selected := make([]incr.Incr[string], 0, len(names))
		for _, name := range names {
			if panel, ok := panels[name]; ok {
				selected = append(selected, panel)
			}
		}
		if len(selected) == 0 {
			return incr.Return(bs, "(no panels)")
		}
		return incr.Map(bs, incr.All(bs, selected...), func(parts []string) string {
			return strings.Join(parts, " ")
		})
	})
	observed := incr.MustObserve(g, dashboard)

	report := func(what string) {
		if err := g.Stabilize(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "%+v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%-28s %-42s open: %s\n", what, observed.Value(), liveFeeds(feeds))
	}

	report("initial layout")

	// Adding a panel opens exactly one new subscription; the existing ones are not
	// disturbed even though the bind rebuilt its subgraph.
	layout.Set([]string{"prices", "trades", "news"})
	report("add the news panel")

	// Removing one closes exactly that subscription.
	layout.Set([]string{"prices", "news"})
	report("drop the trades panel")

	// Replacing the whole layout closes what left and opens what arrived.
	layout.Set([]string{"risk", "positions"})
	report("switch to the risk view")

	// An empty layout releases everything. This is the case that leaks most easily,
	// because there is no remaining node to notice that anything happened.
	layout.Set(nil)
	report("clear the dashboard")

	// And bringing panels back opens fresh subscriptions.
	layout.Set([]string{"prices", "risk"})
	report("restore two panels")

	// Unobserving the whole graph has to release the rest: the dashboard going away is
	// not different, from a resource's point of view, from a panel going away.
	observed.Unobserve(ctx)
	fmt.Printf("%-28s %-42s open: %s\n", "unobserve the dashboard", "(unobserved)", liveFeeds(feeds))

	fmt.Println()
	fmt.Println("per-feed open/close counts, which should be balanced:")
	for _, name := range sortedNames(feeds) {
		f := feeds[name]
		status := "balanced"
		if f.opens != f.closes {
			status = fmt.Sprintf("LEAKED (%d still open)", f.opens-f.closes)
		}
		fmt.Printf("  %-10s opened=%d closed=%d  %s\n", name, f.opens, f.closes, status)
	}
}

// liveFeeds names the currently open subscriptions.
func liveFeeds(feeds map[string]*feed) string {
	var out []string
	for _, name := range sortedNames(feeds) {
		if feeds[name].open {
			out = append(out, name)
		}
	}
	if len(out) == 0 {
		return "(none)"
	}
	return strings.Join(out, ",")
}

func sortedNames(feeds map[string]*feed) []string {
	out := make([]string, 0, len(feeds))
	for name := range feeds {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
