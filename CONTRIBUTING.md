# Contributing

## Running the tests

```sh
go test ./...                # the suite
go test -count 10 ./...      # repeated, which catches order dependence and leaked state
go test -race -count 1 ./... # the concurrency paths: ParallelStabilize, Clock, Var.Set
```

`make ci` runs what CI runs.

## Scaling measurements

Some tests measure elapsed work to assert that an operation has not returned to quadratic
behavior. They are opt-in, because a threshold loose enough to survive a loaded machine
catches nothing:

```sh
INCR_SCALING_TESTS=1 go test -p 1 -count 1 -run 'Test_scaling|_scaling|_work' ./...
```

`-p 1` keeps packages from overlapping, since these compete for cache and GC. What they
assert is also covered by the deterministic work-counting tests, which run normally — a
test that counts recomputes rather than nanoseconds is always the better one to write.

## Fuzzing

Every structural bug found in this library was invisible to a test that checked values:
the values stayed right while the bookkeeping beneath them drifted, and each one needed a
shape nobody had thought to write down. That is what the fuzz targets are for.

| target | package | what it checks |
| --- | --- | --- |
| `FuzzGraph` | root | random graphs stay internally consistent, and unobserving everything drains the graph -- including graphs holding nodes that panic, since recovering from one returns the node to the recompute heap |
| `FuzzMap` | `incrutil/pmap` | a `pmap.Map` against a builtin map, including the cached subtree sizes and heights |
| `FuzzSymmetricDiff` | `incrutil/pmap` | the diff against brute force, with a shared ancestor so the pruning path is taken |

CI runs a short smoke of each, which catches a target that no longer builds or that a
change broke outright. Finding new bugs takes much longer than CI should hold, so run a
real campaign by hand when changing anything structural — edges, heights, teardown,
bind relinking, or the tree's balancing:

```sh
go test -run '^$' -fuzz 'FuzzGraph$' -fuzztime 30m .
go test -run '^$' -fuzz 'FuzzMap$' -fuzztime 30m ./incrutil/pmap
```

`-fuzz` takes one target at a time. A failure writes the offending input to
`testdata/fuzz/<target>/`; commit that file, since it turns the failure into a permanent
regression test that runs with the normal suite.

The two properties `FuzzGraph` asserts are worth preserving as you extend it: that
`ExpertGraph(g).CheckInvariants()` holds after every operation, and that a graph whose
observers have all been released reports zero nodes. The second is what an edge leak
violates, and nothing about a node's value reveals it.

## Linting

```sh
golangci-lint run ./...
```

golangci-lint has to be built with a Go at least as new as the one this module targets, or
it refuses to load — v1.54 fails importing `sync/atomic`, and v1.64 rejects the module
outright. CI pins v2.12.2. The config is v2 format.

## Benchmarks

`_bench/` compares this library against Jane Street's OCaml `incremental`; see
`_bench/README.md` to run it and `_bench/ALGORITHMS.md` for the analysis.

Two things about measuring here, both learned the hard way:

- **Compare interleaved, never across sessions.** This machine drifts by up to 2x between
  runs, which is enough to invent a regression that does not exist or hide one that does.
  `_bench/ab.py 5 "before=./a" "after=./b"` alternates the two and takes per-case minimums.
- **Establish the noise floor before believing a delta.** Build the same source twice and
  A/B the two binaries. It comes out around 1–2%; and adding any code to a hot function —
  even a dead increment — moves the deepest per-node benchmarks by about 5% through code
  layout alone. Deltas that size are not attributable to what the code does.
