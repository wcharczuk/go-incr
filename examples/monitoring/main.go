// Command monitoring is a staleness monitor: several data feeds report heartbeats, and
// the graph decides which are healthy, whether the system as a whole is serving, and
// when to alert.
//
// Anything that watches for something *not* happening depends on time passing rather
// than on an input arriving, which is what makes this kind of code awkward to write and
// worse to test. The usual version sleeps in a loop and reads the wall clock, so its
// tests either take real seconds or are flaky.
//
// Here time is an input like any other. [incr.Clock] is advanced explicitly, so this
// program walks through half an hour of operation in no time at all and would produce
// exactly the same output on any machine, in any order, every run. The same graph in
// production advances the clock from a ticker instead.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/wcharczuk/go-incr"
)

// heartbeat is the last time a feed reported, and what it reported.
type heartbeat struct {
	At    time.Time
	Value int
}

const staleAfter = 5 * time.Minute

func main() {
	ctx := context.Background()
	start := time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)
	clock := incr.NewClock(start)
	g := incr.New(incr.OptGraphMaxHeight(64))

	feeds := []string{"prices", "trades", "reference"}
	beats := make(map[string]incr.VarIncr[heartbeat], len(feeds))
	healthy := make([]incr.Incr[bool], 0, len(feeds))

	// Every interval the clock wakes this node, which is what gives the graph a reason
	// to reconsider staleness when nothing else has changed. Without it, a feed going
	// quiet would produce no event at all and nothing would recompute.
	ticks := incr.AtIntervals(g, clock, time.Minute)

	for _, name := range feeds {
		beat := incr.Var(g, heartbeat{At: start, Value: 0})
		beats[name] = beat

		// A feed is healthy if its last heartbeat is recent. This depends on both the
		// heartbeat and the passage of time, and reads the clock rather than time.Now so
		// that every node in a pass agrees about when "now" is.
		feedHealthy := incr.Map2(g, beat, ticks, func(h heartbeat, _ int) bool {
			return clock.Now().Sub(h.At) < staleAfter
		})
		healthy = append(healthy, feedHealthy)
	}

	// Serving requires every feed. Built on ReduceBalanced, so one feed changing costs
	// O(log n) rather than rereading all of them -- which matters when it is thousands
	// of feeds rather than three.
	serving := incr.ForAll(g, healthy...)

	// Degraded means at least one feed is down, which is a different question and worth
	// tracking separately from "are we serving".
	degraded := incr.Exists(g, incr.Map(g, serving, func(ok bool) bool { return !ok }))

	// A market open time, as a plain fact about the clock rather than something anyone
	// has to remember to poll.
	open := incr.At(g, clock, start.Add(30*time.Minute))

	// What we would page on: down while we are supposed to be serving.
	shouldAlert := incr.Map2(g, degraded, open, func(down, isOpen bool) bool {
		return down && isOpen
	})

	// Capture the state at market open, so that a post-mortem can say what was true then
	// rather than what is true now.
	atOpen := incr.Snapshot(g, clock, serving, start.Add(30*time.Minute), false)

	observedServing := incr.MustObserve(g, serving)
	observedAlert := incr.MustObserve(g, shouldAlert)
	observedAtOpen := incr.MustObserve(g, atOpen)
	observedTicks := incr.MustObserve(g, ticks)

	report := func() {
		if err := g.Stabilize(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "%+v\n", err)
			os.Exit(1)
		}
		state := "serving"
		if !observedServing.Value() {
			state = "DEGRADED"
		}
		alert := ""
		if observedAlert.Value() {
			alert = "  <-- PAGE"
		}
		fmt.Printf("%s  min=%-3d %-9s%s\n",
			clock.Now().Format("15:04"), observedTicks.Value(), state, alert)
	}

	// All feeds fresh at the start.
	beat := func(name string, value int) {
		beats[name].Set(heartbeat{At: clock.Now(), Value: value})
	}
	report()

	// Ten minutes of everything reporting each minute.
	for range 10 {
		clock.AdvanceBy(time.Minute)
		for i, name := range feeds {
			beat(name, i)
		}
		report()
	}

	// The reference feed goes quiet. Nothing signals that; it is the absence of an event.
	// The graph notices anyway, because staleness depends on the clock, and it notices
	// exactly when the deadline passes rather than whenever something else happened to
	// change.
	fmt.Println("\n-- reference feed goes quiet --")
	for range 7 {
		clock.AdvanceBy(time.Minute)
		beat("prices", 1)
		beat("trades", 2)
		report()
	}

	// It comes back. Note the alert clears on the same pass the heartbeat lands.
	fmt.Println("\n-- reference feed recovers --")
	clock.AdvanceBy(time.Minute)
	for i, name := range feeds {
		beat(name, i)
	}
	report()

	// Run up to and past the market open, with the reference feed quiet again, so the
	// alert becomes live only once the market is open.
	fmt.Println("\n-- quiet again, running into the open --")
	for range 14 {
		clock.AdvanceBy(time.Minute)
		beat("prices", 1)
		beat("trades", 2)
		report()
	}

	fmt.Println()
	fmt.Printf("state captured at the open: serving=%v\n", observedAtOpen.Value())

	// And the snapshot does not move afterwards, even as the live state does.
	clock.AdvanceBy(time.Hour)
	for i, name := range feeds {
		beat(name, i)
	}
	if err := g.Stabilize(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
	fmt.Printf("an hour later: serving=%v, but the captured state is still %v\n",
		observedServing.Value(), observedAtOpen.Value())
}
