go-incremental
==============

`go-incremental` is an incremental computation library designed for parallel computation of large graphs of operations.

It is useful in situations where you want to compute efficiently the outputs of computations where only a subset of the computation changes for a given pass.

# Inspiration

The inspiration for `go-incremental` is Jane Street's [incremental](https://github.com/janestreet/incremental) library.

# Core Concepts

A computation is a graph rooted at a given list of outputs. Each output is fed by potentially many parents.

There are a core set of orthogonal node types that are composed into computations:
- Return : effectively a "constant" computation, or just a returned value.
- Var : effectively a "variable" computation, these are the inputs to the computation.
- Map : a computation that takes an input incremental of a given type and a function to modify the value of that input, returning potentially a new type of value, as an incremental of that output type.
- Bind : similar to Map, but dynamic, in that the output of the given function should be a computation and not a value.
- Map2 : combine two inputs into a single output.
- MapIf : a branching computation that returns one of two inputs based on a boolean input incremental computation.
- BindIf : similar to MapIf, but dynamic.

# Example

Let's say that we want to make a convoluted computation that takes an input:

```go
	output := incr.Map2[float64](
		incr.Var(3.14),
		incr.Map(
			incr.Return(10.0),
			func(a float64) float64 {
				return a + 5
			},
		),
		func(a0, a1 float64) float64 {
			return a0 + a1
		},
	)
	_ = incr.Stabilize(context.Background(), output)
	fmt.Println(output.Value()) // prints 18.14
```
