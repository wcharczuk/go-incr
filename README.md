go-incr
==============

[![Continuous Integration](https://github.com/wcharczuk/go-incr/actions/workflows/ci.yml/badge.svg)](https://github.com/wcharczuk/go-incr/actions/workflows/ci.yml)

![Graph](https://github.com/wcharczuk/go-incr/blob/main/_assets/small_graph.png)

`go-incr` is an incremental computation library designed for partial computation of large graphs of operations.

It is useful in situations where you want to efficiently compute the outputs of computations where only a subset of the computation changes for a given pass.

Think Excel spreadsheets and formulas, but in code.

# Inspiration

The inspiration for `go-incr` is Jane Street's [incremental](https://github.com/janestreet/incremental) library.

# Usage

Given an exmaple computation:

```go
ctx := context.Background()
v0 := incr.Var(ctx, "foo")
v1 := incr.Var(ctx, "bar")

output := incr.Map2(v0, v1, func(a, b string) string { return a + " and " + b })
```

In order to realize the values, we need to observe nodes in a graph, and then call `Stabilize` on the graph:

```go
g := incr.New()
o := incr.Observe(ctx, g, output)
if err := g.Stabilize(context.Background()); err != nil {
  // ... handle error if it comes up
}
```

`Stabilize` then does the full recomputation, with the "observer" `o` marking the graph up from the `output` map as observed.

# A word on how to use this library effectively

It can be tempting when looking at what this library can do to say, "We should make each operation in our process incrementally computed", and that would not be the ideal approach, specifically because making a computation incrementally computed adds overhead for each operation.

A more ideal balance is to use a coarse-grain approach to making a computation incrementally computed, specifically breaking up chunks that are mostly atomic, and making those chunks incrementally computed, but each chunk may incorporate multiple operations.

# More advanced use cases

There are simplified versions of common node types (e.g. `Map` and `Map2`) as well as more advanced versions (e.g. `MapContext` and `Map2Context`) that are intended for real world use cases, and facilitating returning errors from nodes.

Errors returned by incremental nodes will halt the computation for the stabilization pass in serial mode completely. In parallel mode, the current height "block" will finish processing in parallel, aborting subsequent height blocks from being recomputed.

# Design Choices

There is some consideration with this library on the balance between hiding mutable implemenation details to protect against [Hyrum's Law](https://www.hyrumslaw.com/) issues, and surfacing enough utility helpers to allow users to extend this library for their own use cases (specifically through `incr.Expert...` types.)

As a result, most of the features of this library can be leveraged externally, but little of the internal state and recomputation mechanism is exposed to the user.

Specific implications of this are, the `INode` interface includes a function that returns the `Node` metadata, but this `Node` struct has few exported fields on it, and users of this library should not really concern themselves with what's on it, just that it gets supplied to `Incr` through the interface implementation.

# Implementation details

Internally `go-incr` uses a pseudo-height recomputation adjacency list similar to the Jane Street implementation. This offers "good enough" approximation of a heap while also allowing for fast iteration between results (e.g. faster than a traditional heap's O(log(n)) performance).

`go-incr` also supports multiple "graphs" rooted in specific variables.

# A word on `Bind`

The `Bind` incremental is by far the most powerful, and as a result complicated, node type in the ecosystem.

To explain its purpose; there are situations where you want to dynamically swap out parts of a graph rooted at a given node. A concrete example might be in a user interface swapping out the displayed controls, which may be made up of individual computations.

The effect of `Bind` is that "children" of a `Bind` node may have their heights change significantly depending on the changes made to the "bound" nodes, and this has implications for the recomputation heap as a result, specifically that it has to handle updating the reflected height of nodes that may already be in the heap.

`Bind` nodes may also return `Bind` nodes themselves, creating fun and interesting implications for how an inner right-hand-side incremental needs to be propagated through multiple layers of binds, and reflect both its and the original bind children's true height through recomputations. To do this, we adopt the trick from the main ocaml library of creating two new pieces of state; a `bind-lhs-change` node that links the original bind input, and the right-hand-side (or output) incremental of the bind, making sure that the rhs respects the height of the transitive dependency of the bind's input. We also maintain a "scope", or a list of all the nodes that were created in respect to the rhs, and when the outer bind stabilizes, we also propagate changes to inner binds, regardless of where they are in the rhs graph. If this sounds complicated, it is!

An example of one such case:

![Bind Regression](https://github.com/wcharczuk/go-incr/blob/main/_assets/bind_regression.png)

Because `Bind` nodes rely on scopes to operate correctly, the bind function you must provide takes a context argument. This context argument should be passed to node constructors within the bind function. This lets us track which nodes were created in the bind scope, helping us maintain height invariants and link nodes correctly.

An example of a use case for bind might be:

```
ctx := context.Background()
t1 := Map(ctx, Return(ctx, "hello"), func(v string) string { return v + " world!" })
t2v := Var(ctx, "a")
t2 := Bind(ctx, t2v, func(ctx context.Context, t2vv string) Incr[string] {
  return Map(ctx, t1, func(v string) string { return v + " Ipsum" })
})
...

```

Here `t1` is _not_ created within a bind scope, but the map that adds `" Ipsum"` to the value _is_ created within a bind scope. This is done transparently by passing the context through the `Map` constructor within the bind.

# Progress

Many of the original library types are implemented, including:
- Always|Timer
- Bind(If)
- Cutoff(2)
- Freeze
- Map(2,3,If,N)
- Observe
- Return
- Var
- Watch

With these, you can create 90% of what this library is typically needed for, though some others would be relatively straightforward to implement given the primitives already implemented.

An example of likely extension to this to facilitate some more advanced use cases; adding the ability to set inputs _after_ nodes have been created, as well as the ability to return untyped values from nodes.
