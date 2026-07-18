// Command floor establishes a lower bound for what a chain of 2048 incremental
// nodes can cost in Go, to attribute the remaining gap to OCaml's incremental
// between "Go's floor for this shape of work" and "overhead go-incr could still
// remove".
//
// Each variant does the same essential per-node work that stabilization must do:
// call the node's function, store the result, stamp two counters, and move to the
// next node. They differ only in how the node is represented.
package main

import (
	"fmt"
	"time"
)

const depth = 2048

// ---- variant A: monomorphic, contiguous -----------------------------------
// The best case Go can offer: nodes in one flat slice, concrete types, a direct
// (non-interface) call per node.

type nodeA struct {
	value        int
	recomputedAt uint64
	changedAt    uint64
	fn           func(int) int
}

func buildA() []nodeA {
	ns := make([]nodeA, depth)
	for i := range ns {
		ns[i].fn = func(x int) int { return x + 1 }
	}
	return ns
}

func runA(ns []nodeA, seed int, stamp uint64) int {
	v := seed
	for i := range ns {
		n := &ns[i]
		v = n.fn(v)
		n.value = v
		n.recomputedAt = stamp
		n.changedAt = stamp
	}
	return v
}

// ---- variant B: interface-dispatched --------------------------------------
// As above, but each node is reached through an interface and its metadata via a
// virtual method, which is how go-incr represents nodes.

type inode interface {
	meta() *metaB
	stabilize(int) int
}

type metaB struct {
	value        int
	recomputedAt uint64
	changedAt    uint64
	next         *metaB
}

type nodeB struct {
	m  metaB
	fn func(int) int
}

func (n *nodeB) meta() *metaB        { return &n.m }
func (n *nodeB) stabilize(x int) int { return n.fn(x) }

func buildB() []inode {
	ns := make([]inode, depth)
	for i := range ns {
		ns[i] = &nodeB{fn: func(x int) int { return x + 1 }}
	}
	return ns
}

func runB(ns []inode, seed int, stamp uint64) int {
	v := seed
	for _, n := range ns {
		v = n.stabilize(v)
		m := n.meta()
		m.value = v
		m.recomputedAt = stamp
		m.changedAt = stamp
	}
	return v
}

// ---- variant C: interface-dispatched, node-sized, scattered ---------------
// As B, but each node carries the same ~400 bytes of metadata a real go-incr
// Node does, individually heap allocated in a shuffled order. This is what
// isolates the cache cost of a large, pointer-rich node reached by pointer.

type metaC struct {
	value        int
	recomputedAt uint64
	changedAt    uint64
	next         *metaC
	// padding to approximate the real Node's footprint; pointer-free so it costs
	// footprint without adding GC scan work beyond the fields above.
	_pad [368]byte
}

type nodeC struct {
	m  metaC
	fn func(int) int
}

func (n *nodeC) meta() *metaC        { return &n.m }
func (n *nodeC) stabilize(x int) int { return n.fn(x) }

type inodeC interface {
	meta() *metaC
	stabilize(int) int
}

func buildC() []inodeC {
	// allocate with gaps so consecutive nodes do not land on adjacent cache
	// lines, the way nodes created across a graph's lifetime would not.
	ns := make([]inodeC, depth)
	var ballast []*[512]byte
	for i := range ns {
		ns[i] = &nodeC{fn: func(x int) int { return x + 1 }}
		ballast = append(ballast, new([512]byte))
	}
	sink = ballast
	return ns
}

var sink any

func runC(ns []inodeC, seed int, stamp uint64) int {
	v := seed
	for _, n := range ns {
		v = n.stabilize(v)
		m := n.meta()
		m.value = v
		m.recomputedAt = stamp
		m.changedAt = stamp
	}
	return v
}

// ---------------------------------------------------------------------------

func bench(name string, op func(stamp uint64) int) {
	for i := 0; i < 5; i++ {
		op(uint64(i))
	}
	batch := 1
	for {
		start := time.Now()
		for i := 0; i < batch; i++ {
			op(uint64(i))
		}
		if time.Since(start).Seconds() >= 0.01 {
			break
		}
		batch *= 2
	}
	best := 0.0
	for round := 0; round < 8; round++ {
		start := time.Now()
		for i := 0; i < batch; i++ {
			op(uint64(i))
		}
		per := time.Since(start).Seconds() / float64(batch)
		if best == 0 || per < best {
			best = per
		}
	}
	fmt.Printf("%-46s %8.1fus/pass  %5.1fns/node\n", name, best*1e6, best*1e9/float64(depth))
}

func main() {
	a := buildA()
	bench("A monomorphic, contiguous", func(s uint64) int { return runA(a, int(s), s) })

	b := buildB()
	bench("B + interface dispatch", func(s uint64) int { return runB(b, int(s), s) })

	c := buildC()
	bench("C + 400B nodes, scattered allocation", func(s uint64) int { return runC(c, int(s), s) })
}
