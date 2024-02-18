`go-incr`
==============

[![Continuous Integration](https://github.com/wcharczuk/go-incr/actions/workflows/ci.yml/badge.svg)](https://github.com/wcharczuk/go-incr/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/wcharczuk/go-incr)](https://goreportcard.com/report/github.com/wcharczuk/go-incr)

![Graph](https://github.com/wcharczuk/go-incr/blob/main/_assets/basic_graph.png)

`go-incr` is an optimization tool to help implement partial computation of large graphs of operations.

It's useful in situations where you want to efficiently compute the outputs of computations where only a subset of the computation changes based on changes in inputs.

Think Excel spreadsheets and formulas, but in code.

# Important caveat

Very often code is faster _without_ incremental. Only reach for this library if you've hit a wall and need to add incrementality.

# Inspiration & differences from the original

The inspiration for `go-incr` is Jane Street's [incremental](https://github.com/janestreet/incremental) library.

The key difference from this library versus the Jane Street implementation is _parallelism_. You can stabilize multiple nodes with the same recompute height at once using `ParallelStabilize`. This is especially useful if you have nodes that make network calls or do other non-cpu bound work.

# Basic concepts

A computation in `go-incr` is composed of three "meta" node types.

- "Input" types that act as entry-points for data that we'll need in our computation (think, raw cell values in Excel)
- "Mutator" types that take other nodes as inputs, and transform those input values into new values (think, formulas in Excel) or even new sub-graph components.
- "Observe" type that indicates which values we care about, and ultimately want to display or return.

"Input" types for `go-incr` are `Var` and `Return` typically.

"Mutator" types for `go-incr` are nodes like `Map` and `Bind`.

The observer node is special in that there is only (1) type of observer, and it has a very specific role and cannot be passed as an input to other nodes.

# Usage

A mini-worked example:

```go
import "github.com/wcharczuk/go-incr"
...
g := incr.New()
v0 := incr.Var(g, "hello")
v1 := incr.Var(g, "world!")
m := incr.Map2(g, v0, v1, func(a, b string) string {
  return a+" "+b
})
om := incr.MustObserve(g, m)
_ = g.Stabilize(context.Background())
fmt.Println(om.Value()) // prints "hello world!"
```

This is not intended as a "real" use case for the library, simply as a worked example to show the syntax.

You can see more sample use cases in the `examples/` directory in this repo.

# API compatability guarantees

As of v1.xxx you should assume that the functions and types exported by this library will maintain forward compatability until some future v2 necessitates changing things meaningfully, at which point we'll integrate [semantic import versioning](https://go.googlesource.com/proposal/+/master/design/24301-versioned-go.md) to create a new `/v2/` package. The goal will be to put off a v2 for as long as possible.

An exception to the above are the `incr.Expert...` functions, for which there are no guarantees and the API may change between refs without notice.

# A word on how to use this library effectively

It can be tempting to, when looking at what this library can do, make every operation in a slow piece of code incrementally computed.

This typically will make performance _worse_, as making a computation incrementally computed adds overhead for each operation.

A more ideal balance is to write your code as normal assuming nothing is incrementally computed, then do a coarsed-grain pass, specifically breaking up chunks that are mostly atomic, and making those chunks incrementally computed.

# Error handling and context propagation

There are simplified versions of common node types (`Map` and `Map2`) as well as more advanced versions (`MapContext` and `Map2Context`) that are intended for real world use cases, and facilitating taking contexts and returning errors from operations.

Errors returned by these incremental operations will halt the computation for the stabilization pass, but the effect will be slightly different based on the stabilization method used.

When recomputing serially (using `.Stabilize(...)`) the stabilization pass will return immediately on error and no other nodes will be recomputed.

When recomputing in parallel (using `.ParallelStabilize(...)`), the current height block will finish stabilizing, and subsequent height blocks will not be recomputed.

# Design Choices

There is some consideration with this library on the balance between hiding mutable implemenation details to protect against [Hyrum's Law](https://www.hyrumslaw.com/) issues, and surfacing enough utility helpers to allow users to extend this library for their own use cases (specifically through `incr.Expert...` types.)

As a result, most of the features of this library can be leveraged externally, but little of the internal state and recomputation mechanism is exposed to the user.

Specific implications of this are, the `INode` interface includes a function that returns the `Node` metadata, but this `Node` struct has few exported fields on it, and users of this library should not really concern themselves with what's on it, just that it gets supplied to `Incr` through the interface implementation.

# Implementation details

To determine which nodes to recompute `go-incr` uses a partial-order pseudo-height adjacency list similar to the Jane Street implementation. This offers "good enough" approximation of a heap while also allowing for fast iteration (e.g. faster than a traditional heap's O(log(n)) performance).

The underlying assumption is that nodes with the same pseudo-height _do not depend on each other_, and as a result can be recomputed in parallel.

# A word on `Bind`

The `Bind` incremental is by far the most powerful, and as a result complicated, node type in the ecosystem.

To explain its purpose; there are situations where you want to dynamically swap out parts of a graph rooted at a given node. A concrete example might be in a user interface swapping out the displayed controls, which may be made up of individual computations.

The effect of `Bind` is that "children" of a `Bind` node may have their heights change significantly depending on the changes made to the "bound" nodes, and this has implications for the recomputation heap as a result, specifically that it has to handle updating the reflected height of nodes that may already be in the heap.

`Bind` nodes may also return `Bind` nodes themselves, creating fun and interesting implications for how an inner right-hand-side incremental needs to be propagated through multiple layers of binds, and reflect both its and the original bind children's true height through recomputations. To do this, we adopt the trick from the main ocaml library of creating two new pieces of state; a `bind-lhs-change` node that links the original bind input, and the right-hand-side incremental of the bind, making sure that the rhs respects the height of the transitive dependency of the bind's input. We also maintain a "scope", or a list of all the nodes that were created in respect to the rhs, and when the outer bind stabilizes, we also propagate changes to inner binds, regardless of where they are in the rhs graph.

If this sounds complicated, it is!

# A word on `Scopes`

Because `Bind` nodes rely on scopes to operate correctly, and specifically know which nodes it was responsible for creating, the function you must provide to the `Bind` node constructor is passed `Scope` argument. This scope argument should be passed to node constructors within the bind function.

An example of a use case for bind might be:

```
g := incr.New()
t1 := incr.Map(g, Return(g, "hello"), func(v string) string { return v + " world!" })
t2v := incr.Var(g, "a")
t2 := incr.Bind(g, t2v, func(scope incr.Scope, t2vv string) Incr[string] {
  return Map(scope, t1, func(v string) string { return v + " Ipsum" })
})
```

Here `t1` is _not_ created within a bind scope (it's created in the top level scope by passing in the `Graph` reference), but the map that adds `" Ipsum"` to the value _is_ created within a bind scope. This is done transparently by passing the scope through the `Map` constructor within the bind.

# Progress

Many of the original library types are implemented, including:
- Always|Timer
- Bind(2,3,4,If)
- Cutoff(2)
- Freeze
- Map(2,3,4,If,N)
- Observe
- Return
- Var
- Watch

With these, you can create 90% of what this library is typically needed for, though some others would be relatively straightforward to implement given the primitives already implemented.

An example of likely extension to this to facilitate some more advanced use cases; adding the ability to set inputs _after_ nodes have been created, as well as the ability to return un-typed values from nodes.
