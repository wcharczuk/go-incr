Changelog
=========

## Unreleased

Modernization pass. The minimum Go version moves to 1.25, which is a breaking change for
consumers; the additions below are otherwise backwards compatible except where noted.

1.25 is the floor because `testing/synctest` needs it, and that is the newest thing this
module uses -- verified by building and testing against it. Requiring 1.26 would cost
consumers a release of headroom for nothing.

### Performance

Incremental updates are roughly 2.5x faster and several pathological cases are gone.
`_bench/ALGORITHMS.md` has the full analysis, including a comparison against Jane
Street's OCaml implementation. The work that mattered:

- The recompute heap searched a height block linearly to remove a node that was already
  intrusively linked, and rescanned from height zero for the next minimum on every
  removal. Both are now O(1) and monotone respectively.
- A dependent whose only input just recomputed is now recomputed directly rather than
  being routed through the recompute heap, as the reference does. A chain of maps no
  longer touches the heap at all.
- The graph's node registry is a slice with each node's position recorded on the node,
  rather than a map keyed by a 16 byte identifier. Insert and remove are both O(1) with no
  hashing, which a bind rebuild pays per node it creates and per node it destroys. Nothing
  ever looked a node up by identifier: `Has` takes the node itself and now answers from a
  flag. Measured interleaved against the map: construction 19-31% faster (`wide/construct`
  at 16384 is 0.69x), bind swaps 7-22% faster, update paths unchanged. The node list's
  index consistency and the node count identity are both checked by `CheckInvariants`, so
  the fuzzer verifies them on every operation.
- Node allocation, which dominates bind-heavy graphs, was reduced by caching optional
  interfaces rather than bound method values, backing the parent and child lists with
  inline storage, and moving rarely used fields behind a lazily allocated side struct.

### Fixed

- **`OnInvalidated` handlers could fire more than once for one transition.** The handlers
  ran before the check for whether the node was already invalid, and a node is reached by
  invalidation more than once -- teardown and a bind's own invalidation walk both reach the
  nodes of a discarded right-hand side. A handler releasing something the node held would
  release it twice.
- **A node whose stabilization failed was never retried.** The stamp marking a node as
  recomputed is applied before its function runs, so a failed node was indistinguishable
  from one that succeeded: no longer stale with respect to its inputs, and not in the
  recompute heap. A transient failure -- a request that timed out, a query that deadlocked
  -- stranded the node's old value permanently while every later pass reported success.
  Failed nodes are now returned to the heap, so a later pass tries them again. This makes
  `OptGraphClearRecomputeHeapOnError(false)` mean what its name says; two tests asserted
  the old behavior and have been updated.
- **Unbounded leak in bind relinking.** `changeParent` removed a child from its old
  parent but not the reciprocal edge, so a bind rewriting its right-hand side left one
  stale parent behind per rebuild — holding every superseded subgraph reachable and
  lengthening every staleness check.
- **Teardown was exponential for duplicate edges.** A node taking the same input in
  several slots (`Map3(x, x, x)`) was torn down once per duplicate edge, re-walking its
  subgraph each time: 0.6ms at depth 8 and 4.0 seconds at depth 16. Also corrupted the
  graph's node count.
- **Quadratic teardown of wide nodes.** Removing the edges of a node with many
  dependents, many inputs, or many observers was O(n) per edge. A wide fan-out took
  197ms to tear down at 8192 dependents; it is now linear in total.

### Added

- `ReduceBalanced`, `UnorderedArrayFold`, `All`, `ArrayFold`, `ForAll`, `Exists`,
  `DependOn`, `Join`.
- `CutoffAlways`, `CutoffNever`, `CutoffEqual`, `CutoffEqualFunc`.
- `VarEqual` and `VarEqualFunc`, vars that ignore being set to the value they already
  hold. A no-op write over 4096 dependents goes from 329us to 87ns, because the graph is
  not consulted at all.
- `VarIncr.Update`, for setting a var from its current value.
- Node lifecycle handlers: `Node.OnBecameNecessary`, `OnBecameUnnecessary`,
  `OnInvalidated`.
- `Clock`, `At`, `AtIntervals`, `Snapshot`, `StepFunction` — an explicitly advanced clock,
  so time-dependent graphs are deterministic and a stabilization sees a single instant. A
  step function is woken once per step rather than continuously.
- `IChildChanged`, which lets a node be told which of its inputs changed instead of
  reading them all. This is what makes constant-time folds possible.
- `IExpertGraph.CheckInvariants`, which verifies that a graph's edges agree with each
  other and that heights are ordered. `IExpertNode`'s one-sided edge methods deliberately
  change one side of an edge, since a caller reconstructing a graph from outside needs
  that; the failure mode is quiet, so this makes it detectable.
- `incrutil/pmap`: an immutable ordered map with structural sharing and a symmetric
  difference proportional to the number of changes.
- `incrutil/mapi`: `MapValues`, `FilterMapValues`, `Merge`, `UnorderedFold`, `Reduce`,
  `MaxValue`, `MinValue`, `Subrange`, `Partition`, `Join`, `Changes`, `Sum`,
  `Cardinality`, `Counti`, `Keys`, and `Selector` — which hands out an incremental per key
  so that one key changing recomputes the consumers of that key rather than every watcher:
  measured over 256 watchers, 1 recompute against 256.
- `pmap.Map.Nth` and `Rank`, positional access in O(log n) using the subtree sizes the
  tree already maintains.
- Fuzz targets, since every structural bug found in this library was invisible to a test
  that checked values -- the values stayed right while the bookkeeping beneath them
  drifted, and each one needed a shape nobody had thought to write down. `FuzzGraph` builds
  random graphs and checks after every operation that the structure is consistent and that
  unobserving everything drains the graph completely, which is the property the bind leak
  violated. `FuzzMap` checks a `pmap.Map` against a builtin map including the cached
  subtree sizes and heights that `Len`, `Nth`, `Rank` and rebalancing all trust, and
  `FuzzSymmetricDiff` checks the diff against brute force with a shared ancestor present,
  so a bug in the pointer-identity pruning shows up as a change silently never reported.
  About 10 million executions found nothing; CI runs a short smoke of each.
- `scaling_test.go`, which asserts growth exponents per graph shape so a pathological
  case fails the build rather than being discovered by a user with a large graph.

### Changed

- A panic in a node's computation is reported as a `PanicError` rather than unwinding
  through the caller. The panic value and the stack are both kept, the node is named, and
  the node is returned to the recompute heap like any other failure. Under
  `ParallelStabilize` this is not a convenience: a panic in a worker goroutine cannot be
  recovered by the caller, so it previously ended the process, and no amount of care on the
  calling side could have prevented that.

  The guard is one recover per pass rather than one per node, which is why it is free. A
  deferred guard around each node measured 2.5ns against roughly 25ns of work, and cost
  17-22% on `wide/update_all` end to end; the node responsible is instead recorded with a
  single store, which measured 0.4ns and 1.00x. Panics raised outside a node's computation
  are reported with no node attributed.
- `Stabilize` and `ParallelStabilize` honor context cancellation. A cancelled pass stops
  and returns the context's cause; nodes not yet recomputed stay in the recompute heap, so
  the state left behind is the same as an error abort and stabilizing again continues from
  where it stopped. A context that cannot be cancelled costs nothing. Serial stabilization
  checks once every 64 nodes, parallel once per height block.
- Minimum Go version is 1.25, which is what `testing/synctest` requires.
- `NewSequentialIdentifierProvier` is renamed `NewSequentialIdentifierProvider`. The
  misspelling remains as a deprecated alias.
- `incrutil/mapi.Added` and `Removed` are deprecated: they compare and clone whole
  builtin maps, so a single key change costs O(n). Use a `pmap.Map` with `Changes`.
- `IExpertNode`'s one-sided edge methods are documented as changing one side of an edge
  only, leaving the caller responsible for the other half and for the heights. That was
  always true; it was not written down, and it is the shape of the relinking leak fixed
  above. They are kept rather than made symmetric because reconstructing a graph from
  outside the library needs each half separately.

### Known gaps

- No graph-wide default cutoff, and there cannot be one: comparing two values requires
  them to be comparable, and an `Incr[A]` holds any type. `VarEqual` and `CutoffEqual`
  carry the constraint at construction instead, which is where the type is known.
- `Incr_map`'s `transpose` and the nested-map `collapse`/`expand` operators are not
  implemented; they are the most niche of that family and nothing here needed them.
