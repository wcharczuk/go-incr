# Algorithmic comparison: go-incr vs. Jane Street `incremental`

Notes from reading `~/workspace/third_party/incremental` (v0.18 preview) against
go-incr, focused on what happens during stabilization.

**Naming is inverted between the two libraries.** In `incremental`, a node's
*children* are its inputs and its *parents* are its dependents. go-incr uses the
opposite convention. This document uses go-incr's vocabulary throughout
(parents = inputs, children = dependents).

## Where the two agree

The core design is genuinely the same, and go-incr's asymptotics already match:

- Height-indexed recompute heap: an array of buckets indexed by node height,
  drained lowest-height-first, with a `height_lower_bound` / `minHeight` cursor.
  Both are O(1) per add/remove, not a comparison heap.
- Intrusive linked lists for the height buckets, threaded through fields on the
  node itself (`next_in_recompute_heap` / `nextInRecomputeHeap`).
- `height_in_recompute_heap = -1` (go-incr: `HeightUnset`) doubles as the O(1)
  "is this node queued" test.
- A separate adjust-heights heap for repairing heights when a bind rewrites
  structure, and invalidation propagated to dependents.
- `needs_to_be_computed = is_necessary && is_stale`, with staleness derived from
  `changed_at` of inputs vs. `recomputed_at` of the node.

So the measured gap was never algorithmic complexity. It was constant factors
plus three specific optimizations `incremental` performs and go-incr did not.

## Fixed in this pass

### 1. O(1) unlink instead of a linear search (algorithmic)

`recomputeHeapList.remove` took an `Identifier` and **scanned the height block**
to find the node, even though the node is intrusively linked and the caller
already held it. `incremental`'s `unlink` just splices `prev`/`next`.

This made every height adjustment O(block size), so on a wide graph where many
nodes share a height, repairing heights degraded toward quadratic. Now
`removeNode(*Node)` splices directly; the id-based variant is kept for callers
that only hold an identifier.

### 2. Monotone min-height scan (algorithmic)

`nextMinHeightUnsafe` restarted its scan **from height 0** on every removal.
`incremental` only ever moves `height_lower_bound` upward while draining.

Draining a heap spanning H heights was therefore O(H²) instead of O(H). It did
not show up in the deep-chain benchmark only because that graph holds one node in
the heap at a time; any graph with many populated heights paid it. Now callers
that empty a block resume the scan above it.

### 3. Recompute a single dependent directly, skipping the heap (algorithmic)

This is the biggest one, and the clearest thing go-incr was missing.
`incremental`'s `maybe_change_value` does not always enqueue a dependent — it
checks `can_recompute_now` and, when true, calls `recompute` on it immediately.
It tracks the frequency in a dedicated statistic
(`num_nodes_recomputed_directly_because_one_child`), which is a good hint at how
much it relies on this.

The reasoning: adding a node to the recompute heap and taking it straight back
out is only necessary if something else must be stabilized in between. For a
dependent whose *only* input is the node that just recomputed, nothing else can
be pending. A chain of maps therefore recomputes with **zero** heap traffic.

go-incr now does the same, with these guards (mirroring the reference):

- the dependent has exactly one input;
- its scope has already stabilized (`parent.height > child.createdIn.scopeHeight()`),
  so a bind cannot still invalidate it;
- it does not itself rewrite graph structure;
- it is not an `always` node, since `Stabilize` re-queues those by inspecting
  what it pulled from the heap, and a chained node never appears there.

Implemented as an **iterative loop** rather than the reference's recursion, so a
long chain costs constant stack instead of one frame per node.

`incremental` has a second, independent bypass for the same purpose: if the
dependent's height is `<=` the recompute heap's current minimum, recompute it
immediately regardless of how many inputs it has, because heights strictly
increase along edges and so any stale input would have to be queued *below* the
heap's minimum — a contradiction. go-incr now does this too, which is what covers
the interior of a `map2` reduction tree.

Making this sound requires the decision to be taken **after** every sibling has
been queued, so the loop holds one child back and evaluates it last. Getting that
ordering right is what exposed the duplicate-child bug below.

### 3a. Latent bug found: a child can appear twice

A node may appear in its parent's `children` more than once — the simplest case
being `Map2(g, x, x)`, where `x`'s children contain that map2 twice. Nothing
prevented such a node from being both held back for direct recomputation *and*
queued in the recompute heap, which left it recomputed while still linked into a
height block; the heap's `count` then disagreed with its lists and the next
`removeMinUnsafe` dereferenced a nil node.

This was latent in the first version of the direct-recompute change and only
surfaced once the loop was restructured. `recompute_direct_test.go` covers it,
along with a test asserting the chain optimization actually engages (a regression
there would otherwise be silent — correct values, just slower).

### 4. Constant-factor work in the hot path

- Recompute-heap lists were threaded through `INode` (interface), making every
  link/unlink field access a non-inlinable virtual `Node()` call. A line-level
  profile showed 90 of 100ms in `push` on two statements that only did
  `v.Node().field = nil`. Now threaded through `*Node`, with a `self INode` field
  so the heap can still hand the interface value back.
- `heights` was `[]*recomputeHeapList`, costing an allocation per height and a
  pointer hop per probe; now a flat `[]recomputeHeapList`.
- The child-propagation loop called `c.Node()` three times and evaluated all
  three predicates eagerly, including the parent-walking `isStale()`. Now hoisted
  and short-circuited cheapest-first.
- `Node` was 408 bytes with hot fields scattered across offsets 80–392, touching
  ~5 cache lines per visit. Reordered hot-first: they now sit in offsets 0–144.
- Two `atomic.AddUint64` statistics counters per node on the serial path, where a
  plain increment is equivalent.
- `addNode` did two map operations on a 16-byte key per node; now one, gated by
  an `inGraph` flag. `zeroNode` hashed identifiers to delete from two maps that
  are empty outside stabilization; now checked first.

### 5. Allocation, for the bind path

Bind-heavy graphs were the last holdout at ~3x, and there the cost genuinely was
allocation and GC rather than scheduling: every time a bind rewrites its
right-hand side it builds a fresh node per node in the new subgraph.

An allocation profile of `bind/wide_swap/4096` was necessary here, because the
answer by *bytes* and the answer by *object count* pointed at different things:

- By bytes, `NewNode` was 63% — so `Node` was shrunk from 408 to 288 bytes by
  moving the fields most nodes never use (label, metadata, sentinels, and the
  three user callback lists) behind a lazily allocated `nodeExtra`. On its own
  this was worth only ~5%: Go's allocator cost is dominated by per-object
  overhead, not size, so trimming bytes without trimming objects does little.
- By object count the picture was different, and more actionable:
  - `remove` (used to unlink parents, children, observers and sentinels)
    **allocated a whole new slice on every call**, even when nothing matched. It
    now compacts in place. Bind teardown calls it per edge.
  - Caching the optional-interface sniffs as **bound method values**
    (`n.stabilizeFn = typed.Stabilize`) allocates a closure per delegate per node
    — up to six allocations for every node created. These now store the
    interface itself (`stabilizer IStabilize`), which allocates nothing and costs
    the same indirect call. This was the single largest win of the three, and it
    helped the update paths too, since less garbage means less GC.

With the closures gone, the next profile put the two graph-bookkeeping slices at
the top: `addChildren` and `addParents` together were **36% of all allocations**,
because a node's `parents` and `children` slices each allocate on first append and
most nodes have exactly one of each. Both now start out backed by a one-element
array stored inside the node (`parentsInline`, `childrenInline`). They stay
ordinary slices, so every reader is unchanged, and growing past one element spills
to the heap on its own since append reallocates when it outgrows its capacity.

The same trick applies to the node types' own input lists, and there were two
variants of the problem:

- `mapIncr` held `parents []INode` initialised with `[]INode{a}` — one allocation
  per node at construction. `map`, `map2`–`map8`, `always` and the bind lhs-change
  node now use fixed `[N]INode` arrays with `Parents()` returning a slice over them.
- Worse, `cutoff`, `cutoff2`, `freeze`, `watch`, `timer` and `map_if` built a
  **fresh slice on every `Parents()` call** rather than once at construction. These
  now fill the same inline array on demand. In a graph using these kinds,
  `Node.nodeParents` was 12% of all allocations; it disappears from the profile
  entirely.

`mapN` keeps a real slice, since its arity is variadic and unbounded.

Measured in isolation on `bind/swap_mixed/128` — a case added specifically to
exercise these kinds, since the rest of the suite only uses `map`/`map2` — this is
worth ~4%, and it is neutral everywhere that does not use them.

Note that `Node.parents`/`children` inline only one element, so a 3-input node
like `map3` still spills to the heap for its graph-side parent list. Widening the
inline arrays would fix that at the cost of bytes on every node; one element was
the right default for these benchmarks but it is a tunable.

Finally, a bind rebuild set `rhsNodes = nil` and regrew it each time, and rebuilt
its main node's input list with a fresh `[]INode{b, rhs}`. The former now
alternates between two buffers (the previous rebuild's list has to stay alive
while its nodes are invalidated, so one spare is the minimum that allows reuse);
the latter writes into a `[2]INode` array.

`Node` ended this pass at 368 bytes, up from the 288 the lazy `nodeExtra` alone achieved (it is 384 today; see section N):
48 bytes for the six interfaces and 32 for the two inline arrays. Every one of
those trades bytes for fewer allocations, and each measured as a win — the
reverse of the intuition that shrinking the struct was the prize. Node *size*
was worth ~5%; node *allocation count* was worth the rest.

### 6. Two pre-existing bugs, one root cause

Both predate this work (verified against pristine HEAD) and both come from the
same place: **a node can appear in a child's input list more than once**, because
a child may take it as several of its inputs -- `Map2(x, x)`, `Map3(x, x, x)`.

`Graph.removeParents` iterated that list and called `removeParent` per entry. But
`unlink` removes edges by identity, so the first call dropped *every* edge between
the pair. The parent was then unnecessary, so it was torn down -- and the second
and third entries found it still unnecessary and tore it down again.

Consequences:

- **Teardown was arity^depth.** Each duplicate edge re-walked the whole subgraph
  beneath it. On a chain of `Map3(cur, cur, cur)`, rebuilding measured 0.6ms at
  depth 8 and **4.0 seconds** at depth 16 -- a ~6500x jump for 2x the depth,
  matching 3^d exactly.
- **`NumNodes()` underflowed.** Each repeat called `removeNode` again, so the node
  counter was decremented once per duplicate edge and wrapped past zero. A plain
  `Map3(m, m, m)` graph reported `18446744073709551612` after unobserving.

The fix is to skip a parent already visited earlier in the list, plus an
idempotence guard in `becameUnnecessary` (a node no longer registered with the
graph has already been walked) as insurance for any other repeated path. Depth 16
went from 4.0s to **38.6us**, and teardown is now linear in subgraph size:

| chain depth | before | after |
|---:|---:|---:|
| 8 | 0.6ms | 17us |
| 16 | 4.0s | 39us |
| 128 | (would not finish) | 291us |

`duplicate_input_test.go` covers both, asserting exact node counts across repeated
bind rebuilds and after unobserving; on the unfixed code they report a drifted
count and the wrapped counter respectively. Cost on graphs without duplicate
edges is ~1-2%, since the dedupe scan runs over input lists of at most a handful
of entries and does nothing at all for single-input nodes.

Worth noting this is the third bug in the same family, after the duplicate-child
hazard in (3a): anywhere go-incr walks an edge list, a node appearing twice is a
case to think about explicitly.

### 6a. Audit of the remaining edge-list walks

Every other place the graph walks an edge list was checked against the same
hazard: invalidation pushing children onto a queue, `becameNecessaryRecursive`
linking parents, the sentinel walk, the recompute child loop, the adjust-heights
walk over children and over a bind's right-scope nodes, `isStaleInRespectToParent`,
`shouldBeInvalidated`, the observer walk, and the diagram renderer.

All of them are safe, but **none of them is safe deliberately** -- each is saved by
an incidental guard: an early return on the first match, a membership flag such as
`heightInRecomputeHeap` or `heightInAdjustHeightsHeap`, an idempotent map write, or
a height comparison that a repeat visit can no longer satisfy. That is why three
bugs of this shape got in: the invariant was never written down or tested.

`duplicate_edge_test.go` drives duplicate-input graphs through each of those walks
-- construction, update propagation, height adjustment under a bind, invalidation
during a bind rebuild, teardown by unobserving, sentinels, parallel stabilization,
diagram rendering and cycle detection -- asserting computed values, an exact and
stable node count, recompute heap self-consistency, and a fully drained graph after
unobserving. Two of those cases fail on the pre-fix code, so the battery has teeth.

### 8. Pathological scaling, hunted systematically

Constant factors cost every user a little; a superlinear cost decides whether the
library is usable past some graph size, and which size that is depends on the
shape rather than on anything the user can see. `scaling_test.go` therefore
measures each graph shape at three sizes and asserts a growth exponent --
log(t2/t1)/log(n2/n1), where 1.0 is linear and 2.0 is quadratic. Thresholds are
loose on purpose: the point is to catch a cliff, not to police a few percent.

Written after the fact, it immediately found four quadratics that the
cross-library benchmarks had missed entirely, because those benchmarks fix a size
and measure a minimum:

| shape | before | after |
|---|---|---|
| teardown of a wide fan-out | 1.96 (8192 deps: **197ms**) | 1.06 (14.5ms) |
| teardown of a wide mapN | 2.01 (8192 inputs: **284ms**) | 1.09 |
| unobserving many observers of one node | 1.98 (4096: **44ms**) | 1.12 |
| removing every input of a wide mapN | 2.05 | 2.03, see below |

The first three share a cause. A node's parent, child and observer lists are
unordered sets searched by identifier on removal. A scan over a short contiguous
slice is right for almost every node, but none of these lists has a bound: one var
can feed thousands of computations, a mapN can take thousands of inputs. Removing
every edge of such a node -- which is what tearing a graph down does -- is then
O(n) per edge and O(n^2) overall. It is teardown rather than steady-state work, so
a user reaches it without asking for anything unusual.

`edge_index.go` gives each of the three lists an identifier-to-positions map once
it grows past 64 entries, making removal O(1). Below the threshold nothing changes
and nothing is allocated. Positions are a list because a node can appear in one of
these lists more than once, when it takes another node as several of its inputs.

One of the four was self-inflicted: `nodeAppearsBefore`, the duplicate-edge dedupe
added in (6), is O(arity^2). That is free for a map3 and quadratic for a mapN with
thousands of inputs -- it accounted for 97% of the wide-mapN teardown profile even
after the edge index went in. `removeParents` now switches to a set past the same
threshold.

The remaining case, removing every input of a wide mapN in a loop, is quadratic and
expected to be: a mapN's inputs are ordered, since the values reach the fold
function in input order, so removing one shifts the rest and n removals move
O(n^2) elements however the entry is located. Unlike the others it is not automatic
work -- it needs an explicit loop over `RemoveInput`. Its threshold is set above 2
with that reasoning recorded, so the test still catches it going cubic. Making it
linear needs either permission to reorder a mapN's inputs, or a tombstoned
representation that preserves order without shifting; both change observable
behavior, so they are decisions rather than fixes.

Two shapes were checked and found already sound: updates under a height ceiling far
above the occupied heights (exponent 0.01, so the recompute heap's scan tracks
occupied heights rather than the ceiling), and cost per rebuild as a long-running
process rebuilds the same bind repeatedly (0.02, which is what (7) fixed).

## Result: equivalent work per stabilization

Both harnesses have a `-stats` mode reporting how many nodes each library
recomputes. This is the algorithmic check, independent of per-node speed: matching
counts mean the two are doing the same work.

| shape | recomputed (go / ocaml) | of which direct (go / ocaml) |
|---|---|---|
| wide/1024, one leaf changed | 12 / 12 | 11 / 11 |
| deep/128 | 129 / 129 | 128 / 128 |
| deep/2048 | 2049 / 2049 | 2048 / 2048 |
| bind/swap_chain/64 | 68 / 68 | 0 / 1 |
| bind/wide_swap/256 | 1280 / 1280 | 0 / 1 |
| wide/1024, no-op write | **12 / 1** | 11 / 0 |

Propagation is now equivalent, and even the scheduling decisions match: the same
number of nodes bypass the heap. Two differences remain, both understood:

- **Bind, off by one.** go-incr conservatively excludes bind main nodes from the
  direct path (see `nodeRequiresHeapOrdering`), where the reference admits them
  with a specific `node.height > lhs_change.height` test. One node out of 68 and
  1280 respectively — not worth the risk of approximating.
- **The no-op write, 12 vs 1.** This is the default cutoff, discussed below. It is
  the one genuine algorithmic difference left.

`initial_recomputed` also differs (2047 vs 3071 on wide/1024) because
`incremental` recomputes var nodes on the first pass and go-incr does not. Values
verified identical; it does not affect update propagation.

## Result: timing

Interleaved measurement, best-of-3, ratio = go-incr / ocaml (lower is better). This is the
result of *this* pass; see "Current standing" below for where things ended up after
everything in this document.

| case | before | after |
|---|---:|---:|
| wide/update_all/1024 | 2.07x | **0.91x** |
| wide/update_all/16384 | 2.32x | **0.95x** |
| wide/update_one/1024 | 3.77x | **1.54x** |
| wide/update_one/16384 | 4.12x | **1.53x** |
| deep/update_one/128 | 4.26x | **1.48x** |
| deep/update_one/2048 | 4.15x | **1.43x** |
| deep/construct/2048 | 3.51x | 1.42x |
| wide/construct/16384 | 2.04x | 1.31x |
| bind/swap_chain/512 | 2.95x | **1.59x** |
| bind/wide_swap/256 | 2.64x | **1.49x** |
| bind/wide_swap/4096 | 2.01x | **1.17x** |
| bind/wide_construct/4096 | 2.04x | 1.28x |
| bind/swap_chain/64 | 3.68x | 2.41x |
| wide/update_same/16384 | 15.42x | 6.20x |

Every case is now within ~1.2–1.6x of the reference except the smallest bind case,
and bulk recomputation of a wide graph is *faster* than it. Since the work counts
match, what remains is per-node constant factor, not algorithm.

`bind/swap_chain/64` at 2.41x is the worst remaining case and the natural next
thing to profile: it is small enough that fixed per-rebuild costs dominate, so it
did not benefit from the allocation work the way the larger bind cases did.

## N. Why single input changes degrade with graph size

Sweeping the wide shape over a 64x range, ratio = go-incr / ocaml:

    n         construct   update_one   update_all
    1024        0.90         1.21         0.87
    4096        0.96         1.23         0.86
    16384       0.91         1.30         0.87
    65536       0.93         1.34          --

Construction and updating every input are **flat** across the whole range, and ahead of the
reference throughout. Only the single-input case degrades, and it does so steadily. Per node
walked, go-incr goes 26.8ns to 28.4ns while the reference goes 22.1ns to 21.2ns -- ours rises
slightly, theirs does not, and the ratio compounds to 1.34.

It is not the amount of work: `update_one` touches one map plus log2(n) tree levels, about 12
nodes at 1024 and 18 at 65536, and the work-count assertions confirm both libraries recompute
the same set. It is where those nodes are. At 65536 the graph is roughly 86MB, and the walk
touches eighteen nodes scattered through it, so nearly every access is a cache miss and often
a TLB miss. The reference's node record is smaller, so the same walk crosses less memory.

Confirmed directly by padding `Node` with 128 bytes and re-measuring the same case:

    update_one/1024     324ns -> 317ns   (no effect)
    update_one/16384    417ns -> 471ns   (+13%)
    update_one/65536    504ns -> 546ns   (+8%)

Footprint has no effect at 1024 and 8-13% at scale, which is the mechanism. Two consequences
worth keeping in mind:

- Per-node footprint is the lever for this gap, not per-node instruction count. `Node` is 384
  bytes; the cold fields still in it are the three edge index maps (24 bytes, nil for all but
  very wide nodes), two diagnostic counters, `kind`, and four `int` fields that would fit in
  `int32`. Roughly 56 bytes are reachable without moving anything hot, which would drop a size
  class. Each such step is worth a few percent *at scale only* -- measured on small graphs it
  looks like noise, which is why section H's padding proxy overstated its own case.
- The slab (section L) groups nodes created together, which is why bulk updates are ahead. It
  cannot help here: the path from one leaf to the root touches one node per tree level, and
  those are far apart in construction order no matter how they are allocated. Co-locating a
  node with its dependents would be a graph-layout problem rather than an allocator one.

## Current standing

Where things ended up after everything in this document, as the agreement of two full runs on
an otherwise idle machine (ratio = go-incr / ocaml, lower is better):

| case | ratio |
|---|---:|
| wide/update_same_equal (`VarEqual`) | **0.35** -- 26ns against 74ns |
| bind/wide_swap/4096 | **0.76 - 0.82** |
| deep/construct/128 | **0.78 - 0.99** |
| deep/construct/1 | **0.86** |
| wide/update_all/16384 | **0.91** |
| wide/construct/1024 | **0.91 - 0.93** |
| wide/update_all/1024 | **0.92** |
| deep/construct/2048 | **0.92 - 0.93** |
| deep/update_one/1 | **0.92 - 0.93** |
| wide/construct/16384 | **0.92** - 1.05 |
| bind/wide_swap/256 | 1.09 |
| bind/swap_chain/512 | 1.14 |
| wide/update_one/1024 | 1.23 - 1.26 |
| bind/wide_construct/4096 | 1.22 - 1.25 |
| deep/update_one/2048 | 1.34 |
| deep/update_one/128 | 1.36 |
| wide/update_one/16384 | 1.37 - 1.38 |
| bind/swap_chain/64 | 1.41 - 1.45 |
| wide/update_same (plain `Var`) | 4.33 - 5.61 |

Allocation-bound work -- construction, bulk updates, swapping a large subgraph -- is ahead of
the reference, from the slab in section L and the closure removal in section I. Fixed per-pass
cost is ahead, from section M. What remains behind is propagating a *single* input change
through many nodes, 1.2 to 1.4x, which is per-node recompute cost: heap ordering, staleness
checks, interface dispatch. `recomputeNode` is 51% flat in a profile of `deep/update_one/2048`
against 11% for the user's own function, and it carries a branch on `parallel` at six points;
specializing a serial version was worth 4-6% (section O), and the rest is footprint
rather than instruction count (section N).

`bind/swap_chain/64` is the smallest bind case and `bind/wide_construct/4096` has swung
between 8.4ms and 11.5ms across variants; treat both as noisy. OCaml's own
`deep/update_one/2048` moved 12% between two runs of this table, which is why these are bands.

## O. (done) Specialize the serial recompute

`recomputeNode` is 51% flat in a profile of `deep/update_one/2048`, against 11% for the node
functions it calls, and it tested the `parallel` flag at six points per node: two counter
increments, the in-flight record, the structural lock around `Stabilize`, the update handler
queue, and the whole children loop. The serial path is now a specialized copy with that flag
resolved away, and `recompute` picks the loop once per chain instead.

Worth 0.94 - 0.97x on every `update_one` case, and neutral elsewhere. Less than the 51% flat
share suggests, because the branches were all perfectly predicted -- what is saved is the
instructions, not the mispredictions.

The cost is a copy of an 85 line function that can drift from its twin. That is guarded by
`Test_serialAndParallelAgree`, which drives five shapes -- deep chain, wide fan-in, duplicated
inputs, a rebuilding bind, nested binds -- through both paths and requires identical values,
node counts and invariants. It was mutation checked: dropping a single `changedAt` stamp from
the serial copy fails it in six places.

## M. (done) Stop reading the clock twice per stabilization

A pass called `time.Now()` in `stabilizeStart`, and then `time.Since` again in `stabilizeEnd`
-- the second one because it was an *argument* to `TracePrintf`, and arguments are evaluated
whether or not the callee does anything with them. So every stabilization read the clock twice
to format a string that was discarded whenever no tracer was attached.

A clock read is 38ns here. An empty stabilization was 86ns, so the two reads were most of it:

    empty stabilization          86ns  ->  22ns
    one node recomputed         149ns  ->  84ns

The reading is now taken only when a stabilization-end handler will receive it or a tracer will
print it, both of which are cheap to test. Against the reference this is what took
`deep/update_one/1` from 1.65x behind to 0.92x, and a rejected `VarEqual` write from 85ns to
26ns. The gain is fixed per pass, so it is largest on the cases that were worst: 0.83x on
`wide/update_one`, 0.31x on `update_same_equal`, 0.97x on `deep/update_one/2048`.

The general lesson is worth more than the fix: a diagnostic that looks free because its output
is discarded is not free if computing the *argument* has a cost. `go vet` will not find these,
and a profile attributes them to the clock rather than to the trace call that asked for them.

## Is Go the limiting factor? No.

`_bench/floor/` measures what a chain of 2048 nodes costs in Go doing only the
essential per-node work — call the node's function, store the result, stamp two
counters, advance — under three representations:

| variant | ns/node |
|---|---:|
| A: monomorphic, contiguous slice | 4.5 |
| B: + reached through an interface, metadata via a virtual method | 5.5 |
| C: + 400-byte nodes, individually allocated with gaps | 5.7 |
| | |
| **ocaml incremental, actual** | **14.8** |
| **go-incr, actual** | **25.0** |

Go's floor for this shape of work is ~5.7ns/node, which is **2.6x below what
`incremental` actually achieves**. Interface dispatch costs about 1ns/node and the
node's size and scatter cost about 0.2ns/node more — together under a fifth of the
remaining gap.

So the remaining ~1.7x is not the language, the GC, interface dispatch, or cache
behaviour. It is roughly 19ns/node of go-incr's own bookkeeping sitting on top of a
4-6ns floor, and `incremental` gets that same bookkeeping down to about 9ns. There
is real headroom left, and it is all in this repository rather than in the runtime.

The profile of `deep/update_one/2048` attributes it roughly as: ~20% the user's own
map function and the indirect call to reach it, ~15% guard evaluation in
`canRecomputeImmediately` and `shouldRecomputeChild`, and the balance the field
writes, counters, cutoff check and error-path setup in `recomputeNode`. Nothing
dominates; it is the accumulation of per-node bookkeeping.

Bind is the exception to all of this. There the cost genuinely is allocation and
GC (~25% of profile in `scanObjectsSmall` / `mallocgcSmallScanNoHeader`), because
every rebuild allocated fresh 400-byte nodes at the time of this measurement -- section L
later removed that, so a steady state rebuild now allocates nothing for node metadata. That
is where OCaml's bump-allocated
minor heap is hardest to match, and why (B) below was the highest-value item at this point.

## Higher-order helpers added

Comparing the two public surfaces, most of what `incremental` offers had an
equivalent here already: `map`/`map2`-`map8`, `bind` and its variants, `cutoff`,
`freeze`, `if_`, `observe`, `save_dot`, the expert interface, and memoization
(`incrutil`'s `BindMemoized` covers `memoize_fun`). The gaps worth closing were the
ones that change asymptotics, because a user who cannot reach for the right
combinator writes the O(n) one instead.

go-incr's only way to combine an unbounded number of inputs was `MapN`, whose
function receives every value, so one input changing costs O(inputs). Three
additions close that:

- **`ReduceBalanced`** combines inputs pairwise through a balanced tree, so a change
  recomputes only the path from the changed leaf to the root: **O(log n)**. It needs
  associativity but not an inverse, so it covers min, max and concatenation. Inputs
  are combined in the order given, so non-commutative operations are safe.
- **`UnorderedArrayFold`** keeps an accumulator and adjusts it from the changed
  input's old and new values: **O(1)**. It needs an inverse -- for a sum,
  `acc - old + new` -- which is exactly the case `ReduceBalanced` cannot beat.
- **`All`** collects inputs into a slice. This is O(n) by construction, since the
  slice has to be built, and exists for convenience; its doc points at the other two
  when the goal is an aggregate rather than the values.

Measured cost of changing one input, asserted in `scaling_test.go`:

| inputs | MapN | ReduceBalanced | UnorderedArrayFold |
|---:|---:|---:|---:|
| 256 | 1236ns | 319ns | 172ns |
| 1024 | 4430ns | 378ns | 178ns |
| 4096 | **17169ns** | **427ns** | **188ns** |
| growth exponent | 0.98 | 0.09 | 0.04 |

At four thousand inputs that is 40x and 91x. The exponents are asserted rather than
the times, so the guarantee survives a change that makes them slower but not
pathological.

### The core hook this needed

An O(1) fold has to know *which* input changed and what it changed from; a node that
reads all its inputs cannot beat O(n) however it is written. `incremental` handles
this by special-casing `Unordered_array_fold` inside `maybe_change_value`, which
calls `child_changed` on the fold as the change propagates.

go-incr now has the general form of that: an `IChildChanged` interface, sniffed once
at initialization like the other optional interfaces and cached on the node, so the
recompute loop notifies via a nil check rather than a type assertion per input
visit. Measured against the previous build across the update and deep benchmarks it
costs 0.98x-1.05x, i.e. nothing outside noise, and `-verify` and `-stats` output are
unchanged.

### Incremental maps

`Incr_map` is a separate Jane Street library and arguably the most used part of that
ecosystem. Its value is that a map-valued computation updates in proportion to the
keys that changed rather than the map's size, and all of it rests on
`Map.symmetric_diff`: on OCaml's persistent balanced maps the old and new maps share
structure, so a diff walks only the subtrees that differ and dismisses identical ones
in constant time by physical equality.

go-incr had `incrutil/mapi.Added` and `Removed`, which are the O(n) trap rather than
the incremental thing -- each pass scans both maps and clones the new one:

| keys | cost of adding one key |
|---:|---:|
| 256 | 17748ns |
| 1024 | 41460ns |
| 4096 | **169302ns** |

Go's builtin map cannot do better. A hash table shares no structure with the map it
was copied from, so comparing two is inherently O(n), and nothing is immutable, so
remembering the previous state means copying it. The gap is a data structure.

**`incrutil/pmap`** supplies the missing structure: a generic immutable ordered map
(`cmp.Ordered` keys) as a persistent AVL tree, so an update rebuilds only the path to
the changed key and shares everything else. `SymmetricDiff` returns an
`iter.Seq[Change[K, V]]` and prunes by pointer, giving the property the whole thing
exists for:

| entries | cost of diffing 8 updates |
|---:|---:|
| 1024 | 421ns |
| 8192 | 667ns |
| 65536 | **734ns** |

Growth exponents of 0.22 and 0.05 -- logarithmic, not linear. `FromGoMap` and
`ToGoMap` bridge to builtin maps at the edges; `FromGoMap` sorts keys first so the
tree shape does not depend on Go's randomized iteration order.

**`incrutil/mapi.MapValues`** is the first operator built on it, the equivalent of
`Incr_map`'s `mapi'`: it diffs its input against the previous pass and applies the
function only to keys that were added or changed, carrying the rest of the output
forward untouched. Changing one key of a 65536-entry map costs ~1.4us and does not
grow with the map, and re-setting an identical map recomputes nothing.

Correctness rests on property tests rather than examples, since a persistent balanced
tree has more rebalancing cases than spot checks reach: 4000 random operations against
a builtin map with tree invariants checked throughout, the diff cross-checked against a
brute-force diff over 200 randomized trials, and an explicit assertion that rebinding
one key of a 1024-entry map rebuilds only the path.

Three more operators follow, all the same shape -- diff the input against the
previous pass, touch only what changed:

- **`FilterMapValues`** (`filter_mapi'`): [MapValues] with the option to drop a key.
  A value whose predicate result flips is added to or removed from the output, so the
  output tracks the predicate as well as the values.
- **`UnorderedFold`** (`unordered_fold`): folds a map to a single value with `add`
  and `remove`, so a changed key withdraws its old contribution and applies its new
  one. This is the keyed counterpart of `UnorderedArrayFold` and carries the same
  requirement of an inverse.
- **`Merge`** (`merge`): combines two incremental maps, diffing both and recomputing
  the union of their changed keys. `MergeElement` reports which sides hold a key.

Each is asserted two ways. `Test_operators_work` counts callbacks and requires
**exactly one per changed key at both 1024 and 65536 entries** -- the precise
algorithmic claim, immune to machine effects. `Test_operators_scaling` then bounds
wall clock, loosely and deliberately: at the largest size the tree spans several
megabytes, so walking it misses cache in a way that grows with size without any extra
work being done, and an early attempt at a tight bound flagged that rather than a
real regression. The work count is the statement; the timing is a backstop.

Correctness is cross-checked against recomputing from scratch -- 500 mutations for the
fold, 400 for the merge with both inputs moving independently -- because a wrong
`add`/`remove` pair or a missed key drifts silently rather than failing.

`Join` is ported in both forms, and the machinery it needed turned out to already be
present -- `MapN.AddInput`/`RemoveInput` does exactly this dynamic linking, via
`Graph.addChild` to link and adjust heights and `checkIfUnnecessary` to release.

- `incr.Join` flattens an `Incr[Incr[A]]`, and is `Bind` with an identity function.
- `mapi.Join` turns a map of incrementals into an incremental map. The outer map's
  diff drives one link or unlink per structural change, and `IChildChanged` -- added
  earlier for `UnorderedArrayFold` -- reports which inner incremental moved, so a value
  change rewrites one entry. Asserted: changing one inner input recomputes exactly one
  inner node whether the map holds 64 entries or 2048.

One subtlety worth recording. An inner incremental that nothing needed until now has
never been computed, so its value is not yet meaningful when the join links it.
Linking makes it necessary and schedules it, but the join has already recomputed in
that pass, so it does not look stale with respect to a parent that changes during the
same pass and would never run again to collect the value -- the first version silently
produced a zero. The join now marks itself stale after linking, and heights guarantee
the second pass runs after the inner node.

The old `Added` and `Removed` still operate on builtin maps and are still linear.
They now sit beside a fast path with nothing warning a caller, so they should be
reimplemented over `pmap` or documented.

### Node lifecycle (done)

`incremental`'s update callbacks carry a four-way state, and go-incr's carry none of
it:

    type 'a Update.t =
      | Necessary of 'a
      | Changed of 'a * 'a   (* old value, new value *)
      | Invalidated
      | Unnecessary

`Node.OnUpdate` takes `func(context.Context)` and `ObserveIncr.OnUpdate` gets the new
value, so none of the transitions were visible. Three are now: `OnBecameNecessary`,
`OnInvalidated` and `OnBecameUnnecessary`. The last is what a caller needs to release
something a node was holding -- a subscription, a file -- and there was previously no
way to hear about it at all.

They take no context, unlike the other handlers. These are structural transitions
reached from paths that carry none, and the alternative was either threading a context
through a wide fan of internal callers or inventing one at the call site; a handler
needing a context can capture it. They fire as the transition happens rather than
being deferred to the end of stabilization, because releasing a resource late is the
wrong behavior, and they must not modify the graph. The handlers live in `nodeExtra`,
so a node that registers none still allocates nothing -- asserted in
`lifecycle_test.go`.

Still not exposed: the **previous** value on a change. A caller wanting it can cache
it in the handler, so this is a convenience rather than a capability.

### Map operators over `pmap` (done)

Built on the diff, each maintained rather than recomputed:

- `Cardinality`, `Counti`, `Sum` -- folds over the change set. `Cardinality` passes no
  equality function, since a rebound value cannot move a count and so should not
  trigger any work.
- `Keys` -- ordered key slice. Honestly O(n) per change, and documented as such: a
  sorted slice cannot be maintained incrementally, so this is convenience only.
- `Subrange` -- a window over a sorted map, with **incremental bounds**. This is the
  point of having an ordered structure. When the map changes it applies only the
  changes falling inside the window, O(changes); when the bounds move it rebuilds from
  a bounded walk that never visits a subtree outside them, O(window). Neither cost
  tracks the size of the collection, which is what makes scrolling a large collection
  viable.
- `pmap.Map.Min`/`Max`/`Range` -- the ordered-tree primitives these rest on.

### Aggregates with no inverse (done)

`UnorderedFold` maintains a running total by withdrawing an entry's contribution and
applying its replacement, which only works when the operation has an inverse. A
maximum does not: withdrawing the current maximum says nothing about the next one.

`pmap.Reducer` handles that case, and the shape of the solution comes from the
representation rather than from new graph machinery. The map is already a balanced
tree whose subtrees are shared between versions, so folding over the tree and
memoizing each subtree's result on the subtree itself means an unchanged subtree is
found in the memo and its walk stops there. A changed key only invalidates the
subtrees on its path to the root.

Measured combines for one changed key:

| entries | combines | full fold would be |
|---:|---:|---:|
| 1024 | 18 | 1023 |
| 8192 | 24 | 8191 |
| 65536 | **30** | 65535 |

That is about 2 x log2(n), an exponent of 0.12 against map size, and ~2000x less work
than folding from scratch at the largest size.

Two details the tests pin down. Subtrees are combined in key order, so an associative
but non-commutative operation -- string concatenation -- is well defined, not merely
lucky. And the memo needs pruning: it holds an entry per subtree and nothing reports
when a version is discarded, so it is rebuilt from the reachable set once it exceeds a
multiple of the map's size, which amortizes against the growth that triggered it.
`Test_Reduce_memoIsBounded` drives 4000 updates and asserts it stays bounded.

`mapi.Reduce` exposes this as a node, with `MaxValue` and `MinValue` over it returning
an `Optional` so an empty map reports absence rather than a zero value.

### Still missing (all since implemented)

This section listed `array_fold`, `for_all` / `exists`, `join`, `depend_on`, the named
cutoffs, `Var.replace ~f`, and -- called out as "the largest thing absent" -- the virtual
clock: `at`, `at_intervals`, `snapshot`, `step_function`, `advance_clock`.

All of them now exist: `ArrayFold`, `ForAll`, `Exists`, `Join`, `DependOn`, `CutoffAlways`,
`CutoffNever`, `CutoffEqual`, `VarIncr.Update`, and `Clock` with `At`, `AtIntervals`,
`Snapshot`, `StepFunction` and `Clock.Advance`. Kept as a record of what the gap analysis
found, not as a to-do list.

The only operators still absent are `Incr_map`'s `transpose` and the nested-map
`collapse`/`expand`, deliberately: they are the most niche of that family and nothing needed
them.

## Go toolchain

`go.mod` was on `go 1.21`, which left two things on the table; it is now `go 1.25`,
a floor that still supports consumers a couple of releases back.

- **Per-iteration loop variables** (1.22). Previously the `x := x` idiom inside a
  loop was load-bearing. The library had no live capture bug, but the footgun was
  armed; the benchmark harness's redundant copies are now removed.
- **`testing.B.Loop`** (1.24), now used throughout `bench_test.go`. It excludes
  setup from the timed region without an explicit `ResetTimer` and keeps the loop
  body from being optimized away.

Two features arrive for free with a recent toolchain and are already visible in
profiles: Swiss maps (`maps.ctrlGroup.matchH2`) and the Green Tea GC
(`tryDeferToSpanScan`, `scanObjectsSmall`).

**PGO** measures at a uniform 3-7% across every case, including on the
interface-heavy hot path where profile-guided devirtualization might have been
expected to do more. It is not something this repository can ship, since PGO
applies to the final binary -- it is worth mentioning to users building
go-incr-heavy binaries, and nothing more.

## Individual investigations

Each item below is lettered, and the letters are historical rather than ordered -- they were
assigned as the questions came up, and later ones were appended where they were written rather
than sorted in. Some of the earliest are now done, and a few turned out to be the wrong idea,
which is worth as much as the ones that landed.

| | | where |
|---|---|---|
| A | Cutoff by default | **decided** -- impossible graph-wide; `VarEqual`/`CutoffEqual` instead |
| B | Shrink `Node` | **done**, and see N for what remains |
| C | Inline the first input and dependent | **done** |
| D | Edges keyed by identity, not input slot | open |
| E | Admit bind main nodes to the direct path | open |
| F | Avoid the redundant `isStale()` | **done** |
| G | Kind-directed dispatch | open, and probably not worth it |
| H | Embed `Node` by value | **tried, reverted** -- 1.9x worse on the widest update |
| I | What actually reduces allocation | findings; three attempts, one winner |
| J | Manage our own node memory? | superseded by L |
| K | Defer teardown to after the pass | **declined** -- 4-6% and it retains garbage |
| L | Slab allocation for node metadata | **done** |
| M | Stop reading the clock twice per pass | **done** -- above, before "Is Go the limiting factor" |
| N | Why single input changes degrade with size | analysis; above, before "Current standing" |
| O | Specialize the serial recompute | **done** -- above |

M, N and O sit earlier in this file rather than in the block below, next to the measurements
they explain.

### A. (decided) Cutoff by default

`incremental`'s `maybe_change_value` compares the new value against the old with the node's
cutoff -- **physical equality by default** -- and if unchanged does not touch `changed_at` and
does not enqueue dependents at all. Propagation stops dead. go-incr had no default, so a write
of an unchanged value propagated through the whole graph: the `wide/update_same` case.

**Resolved: there cannot be a graph-wide default.** Deciding whether two values are equal
requires them to be comparable, and an `Incr[A]` holds `any`. There is no way to express the
constraint at the graph level, so it has to appear where the type is known -- at construction.
`VarEqual` and `VarEqualFunc` were added for the input case, `CutoffEqual` and
`CutoffEqualFunc` for computed values, plus the degenerate `CutoffAlways` and `CutoffNever`.

With `VarEqual` a no-op write costs 26ns against the reference's 74ns, so this case went from
behind to ahead. Through a plain `Var` it remains 4.3 - 5.6x, which is the cost of the default
and is now a documented choice rather than an open question.

### B. (done) Shrink `Node` and cut allocation per node

`Node` was ~400 bytes. Bind rebuilds allocate one per node created, and the bind benchmarks
were dominated by allocation and GC (~25% of profile in `scanObjectsSmall` /
`mallocgcSmallScanNoHeader`). This is where OCaml's bump-allocated minor heap and cheap
generational collection are hardest to match.

What landed, across this pass and later ones: the rarely used fields moved behind a lazily
allocated `nodeExtra`; the optional-interface sniffs became cached interface values rather
than bound method values, which had allocated a closure per delegate per node; and later the
two delegates consulted only while invalidating stopped being cached at all, since asserting
them at the point of use costs less than two words on every node (section N).

`Node` is 384 bytes today. The cold fields still in it are the three edge index maps (nil for
all but very wide nodes), two diagnostic counters, `kind`, and four `int` fields that would fit
in `int32` -- roughly 56 bytes reachable without touching anything hot, which would drop a size
class. Section N measures what that is worth and why it only shows up on large graphs.

Allocation per node was addressed separately and more effectively by section L, which carves
metadata from per-scope slabs.

### C. (done) Inline the first input and dependent

Done throughout; see (5) above. No `[]INode{...}` literal remains in the library.

### 7. A leak in bind relinking, found by asking the reference's question

Investigating (D) below meant asking whether go-incr's parent and child lists are
a bijection, since the reference's O(1) edge removal depends on one. They are not,
and the reason is a bug rather than a design choice.

`Graph.changeParent` removed the child from the old parent's child list but never
removed the old parent from the child's parent list. Linking is symmetric, so the
missing half accumulates: a bind rewriting its right-hand side goes through here on
every rebuild, and each rebuild left one stale parent behind that nothing would
ever remove.

Measured on a single bind rebuilt repeatedly, before the fix:

| rebuilds | len(main.parents) | per rebuild |
|---:|---:|---:|
| ~600 | 801 | 2227ns |
| ~2600 | 3001 | 1759ns |
| ~7600 | 8201 | 1829ns |

The list grows without bound, holding every superseded subgraph reachable, and
every staleness and invalidation check walks it. After the fix the list stays at 2
-- a bind main's two real inputs, its lhs-change node and its current right-hand
side -- and a rebuild costs ~1100-1500ns.

The fix is for `changeParent` to call `Graph.unlink`, which removes both halves,
rather than only the child side. `bind_relink_test.go` asserts the input list stays
bounded across 50 rebuilds and that the whole graph is free of one-sided edges; both
cases fail on the pre-fix code.

On the cross-library bind benchmarks the fix is not a clean win, and the pattern is
reproducible over five interleaved rounds rather than noise: `swap_chain/64` and
`wide_swap/4096` improve to 0.86x and 0.91x, `wide_construct/4096` is unchanged,
and `swap_chain/512`, `wide_swap/256` and `swap_mixed/128` regress to 1.10-1.14x.
The extra half of the unlink is genuinely more work per rebuild. `-stats` output is
identical before and after, so scheduling did not change; I have not fully
accounted for the size of the regression on those three shapes.

The trade is still clearly right -- an unbounded leak is not something to keep for
a few percent -- but it is a trade, not a free improvement. Note also that neither
these benchmarks nor `Benchmark_Stabilize_nestedBinds` rebuild often enough to show
the leak's real cost, which only appears once a process has been running a while;
that is worth remembering when reading the numbers in (D).

### H. (tried, reverted) Embed `Node` by value in the node types

Each node is two allocations: the concrete node struct (`mapIncr` and friends, ~56 bytes)
and the `Node` metadata it points at (384 bytes). Making the metadata a value field
collapses those into one, and construction is allocation-bound, so this looks like a clear
win. In isolation it is: modelling both layouts over a chain of 1024, allocations halve and
construction runs 18% faster.

Implemented across all 29 node types, it is not a win. Measured interleaved:

    bind/wide_construct/4096    0.79x
    deep/construct/128          0.80x
    bind/wide_swap/256          0.86x
    wide/update_one/16384       0.85x
    wide/update_all/16384       1.43x - 2.07x   <-- 

`wide/update_all` at 16384 regresses consistently. Eight alternating runs of the one case,
no overlap between the distributions:

    pointer   3.3  3.8  5.0  5.1  5.2  5.5  5.8  6.3 ms   (median ~5.1)
    embedded  7.2  7.5  8.6  9.7 10.0 10.1 10.7 12.3 ms   (median ~9.8)

It is not the allocator or the collector: `gctrace` shows four cycles and the same heap
sizes for both. It is the access pattern. Leading with 384 bytes of metadata pushes each
node's own fields -- the input, the function, the value, the inline parent array, which is
what a recompute reads -- past six cache lines from the node's address. Putting the
embedded field last instead recovers most of the single-node paths (`update_one/16384`
1.18x to 0.85x) but `update_all` stays 1.43x: that case walks 32k nodes and is the largest
working set in the suite, right at the cache boundary, so it is the one shape where the
layout change dominates.

Reverted. Construction is a one-time cost and the widest update is the steady-state hot
path; trading a ~1.9x regression there for 20% off construction is the wrong direction, and
a pathological case at scale is worse than a slow constructor. `NewNode` keeps its two
allocations deliberately.

The general lesson, which cost a 29-file refactor to learn: a struct-layout model that
measures allocation in isolation predicts the allocation change correctly and tells you
nothing about the traversal. Both halves have to be measured on the real graph.

### L. (done) Slab allocation for node metadata

Sections I and J concluded against managing node memory, on two measurements that turned out
to bound the wrong things. `GOGC=off` bounds "never reclaim", which is not what an allocator
does. And the pooling prototype used a LIFO free list of individually allocated nodes, which
*scatters* them -- the opposite of what an allocator should do.

Carving nodes from contiguous chunks does the opposite of both: reclamation still happens,
just at chunk granularity, and nodes created together are adjacent. Measured on an idle
machine, min chunk 2 doubling to 256:

    bind/wide_swap/256          0.77x
    bind/swap_chain/512         0.80x
    bind/swap_mixed/128         0.80x
    bind/swap_chain/64          0.81x
    bind/wide_swap/4096         0.81x
    wide/construct/16384        0.89x
    wide/update_all/16384       0.90x
    bind/wide_construct/4096    0.97x
    deep/construct/128          1.10x
    deep/construct/1            1.12x

Note `wide/update_all` at 0.90x: an update-path win, which neither embedding (H) nor pooling
(I) managed. It is the contiguity, not the cheaper allocation.

Three things had to be right, and each was found by a failure:

**Reuse, not release.** Dropping a scope's chunk on each rebuild and letting the collector
take it measured **1.65x** on `bind/wide_swap` -- a wide right-hand side then allocates ~1.5MB
of cold memory per rebuild and discards as much, where one-at-a-time allocation recycles the
same warm memory. Resetting the slab and reissuing the slots is what turns this from a
regression into 0.79x.

**Double buffering.** Resetting the slab at the start of a rebuild reissues slots that the
previous generation is still using: teardown happens after the swap, not before, so the old
nodes are live while the replacement is built. This broke nested binds outright. The slots
safe to reissue belong to the generation *before* the one being replaced, which is exactly
why `rhsNodes` and `rhsNodesSpare` already alternate; the slab alternates with them.

**A tiny minimum chunk.** Every bind scope owns a slab, so a graph with thousands of binds
whose right-hand sides are one or two nodes pays the minimum thousands of times. At 8 slots
that cost 27% on `bind/wide_construct/4096`; at 2 it costs nothing, and larger scopes reach a
useful size within a few doublings.

Retention over 20,000 rebuilds is 19KB, against 14KB for one-at-a-time allocation and 2509KB
for a naive shared slab with no reuse -- that last one is what happens when chunks mix
generations, since one surviving node pins its whole chunk.

The remaining cost is deep construction at 1.10x: a small chain pays the growth ramp where
one-at-a-time allocation reuses a hot span. It is a few microseconds on a 33us operation, and
the trade is deliberate -- the wins are on the shapes that get large.

### J. Should this library manage its own node memory?

The case for it: Go's collector is tuned for latency, and a big graph churn wants
throughput. `Stabilize` already blocks, so in-pass latency is not what we are protecting --
what we would want is for the churn's garbage not to cost the surrounding process.

Measured against that goal directly, taking the minimum of three runs per setting:

    setting      bind/swap_chain/512   bind/wide_swap/256   wide/construct/16384
    GOGC=100          179us                 218us                 16.2ms
    GOGC=400          160us  0.89x          191us  0.88x          15.3ms  0.94x
    GOGC=off          209us  1.16x          241us  1.11x          20.5ms  1.26x

**Disabling the collector is worse than leaving it on**, by 11% on bind churn and 26% on
construction. Nothing is being collected, so nothing is being reused, so the heap grows and
the working set stops fitting; `wide/construct` has the largest working set in the suite and
takes the worst of it. The collector's recycling is buying locality, which is the same force
that made embedding the metadata (H) and recycling nodes (I) lose.

That bounds an allocator that never returns memory, which is not the interesting design; see
section L, where chunk-granularity reclamation with reuse does win. What this measurement
does establish is that the collector's recycling is buying locality, so any scheme has to
preserve it -- which is why the naive free list lost and the contiguous slab did not.

What does work is tuning the collector's frequency rather than removing it: `GOGC=400` is
worth 11-12% on churn and 6% on construction, and it is a caller-side line
(`debug.SetGCPercent(400)`, ideally with `debug.SetMemoryLimit` to bound the resulting peak)
that needs nothing from this library. Reaching for it is reasonable for a churn-heavy
workload; a library should not set it on a caller's behalf.

The prior art is also empty. OCaml `incremental` has no pool, no free list, and no arena
anywhere in its source; node records are ordinary allocations reclaimed by the runtime, and
`Node_id.next()` being monotonic is itself evidence that records are never reused. It makes
no calls into `Gc` beyond attaching observer finalizers, and there is no comment anywhere in
its source discussing minor-heap behavior, pauses, or stabilization latency. Its memory
strategy is a structural invariant instead: dependent edges exist only while a node is
necessary, so a subgraph going unnecessary severs everything that could pin it. That is
worth more than any allocator, and it is portable.

### K. Deferring teardown to after the pass

Also considered: reap nodes after a stabilization rather than during it. Measured share of a
`bind/swap_chain/512` swap:

    becameUnnecessary (unlinking)     13%   must stay inline; semantics
    removeNode + zeroNode              4-6% deferrable
    allocator and collector           26%   not addressed by deferring

Only the scrub and registry removal can move, and they are 4-6% of the pass. Deferring them
also *retains* the dead nodes until the queue drains, so peak memory rises and collection is
delayed -- the opposite of the goal. Meanwhile the 26% that actually lands inside the pass is
allocator and collector work, which deferring does not touch.

OCaml does not defer any of this either. Its teardown is entirely eager and inline inside the
recompute loop; the only things it defers to the end of a pass are user-visible callbacks,
`Var.set` calls that arrive during stabilization, observer registration, and weak-table
compaction. The one work queue that looks like batching, `propagate_invalidity`, exists purely
because a parent removing itself would invalidate the iteration over parents -- the same
problem a Go port hits when a dependent removes itself from a slice being ranged over.

### I. What actually reduces allocation

Construction profiles at roughly half allocation and collection, so reducing allocation
looks like the obvious lever. Three attempts at it and one thing that worked:

**Pooling nodes (measured, ~12-17% on churn, not landed).** A free list that recycles the
`Node` metadata on teardown. Measured against the base, with an unsafe prototype that
verifies identical values:

    bind/wide_swap/256     0.83x
    bind/swap_chain/512    0.84x
    bind/swap_mixed/128    0.87x
    construct (all)        0.96 - 1.04x   (nothing to recycle)
    update paths           1.04 - 1.18x   (unattributable, see below)

The bind win is real and outside the noise band. The apparent update cost is not
attributable: the prototype does a redundant 416-byte write per node (`new(Node)` zeroes it,
then the composite literal writes it again), and adding code to `node.go` shifts layout,
which is independently worth ~5% on these cases. A clean prototype would be needed to know.

It is also not safe as written. A caller holding an `Incr` from a discarded bind scope would
observe another node's metadata, and could link it back into the graph. Landing it needs an
opt-in graph option, a generation counter on `Node`, and a checked mode that panics on
access to a recycled node -- and it should be weighed against simply not churning the graph,
which is an order of magnitude better (below).

**Memoizing the bind (measured, slower).** `incrutil.BindMemoized` caches right-hand sides
by key, so a bind over a small key space stops allocating after warmup. Over 4 keys:

    plain bind      11.5us   73 allocs   (before the slab; 9.1us and 34 today)
    memoized bind   15.3us    6 allocs

92% fewer allocations and 33% slower. The cached subgraph still has to be unlinked,
relinked, and walked for necessity and heights on every swap, and that bookkeeping is what
a swap actually costs. Useful when the rebuild function itself is expensive; not a way to
make swaps cheaper.

**Embedding the metadata (see H above, reverted).** Halves the allocation count and
regresses the widest update path ~1.9x.

**What worked: removing an allocation without moving anything.** `Map` and `Map2` allocated a
closure per node purely to adapt their plain function to the context signature. Giving each its
own node type
removed that allocation and touched no layout: construction 7-19% faster, updates unchanged,
no case worse. That is the whole pattern -- three schemes that consolidated or relocated
memory all lost more than they gained, and the one that deleted an allocation outright won
cleanly.

**Node size, and a proxy that overstated itself.** Growing `Node` by 96 bytes of padding
cost 13% on construction and 31% on `wide/update_all/16384`, which made shrinking it look
like the best remaining lever -- and unlike embedding or pooling it removes bytes without
relocating anything. Removing 32 bytes (the two invalidation-only delegates, 416 to 384,
one size class down) gained about 7% on small construction and nothing measurable
elsewhere. Real, worth keeping, and much smaller than the proxy implied: the relationship is
not linear, so a padding experiment bounds the direction but not the magnitude. Reaching the
352 class needs another 32 bytes, which would mean moving the edge indexes behind `ext` and
paying a branch per edge on a construction-hot path; not obviously worth it at this return.

**And the thing that dwarfs all of it: do not churn the graph.** The same computation, once
as a bind that rebuilds its right-hand side and once as a fixed shape driven by cutoffs:

    bind rebuild    11.5us    73 allocs   (before the slab; 9.1us and 34 today)
    stable shape     0.83us    1 alloc

14x and 73x. No allocator this library could ship competes with restructuring the graph, so
the highest-value guidance is a documentation matter, not an allocator one.

### D. Edges keyed by identity, not by input slot

This is the one structural difference left between the two libraries, and it is
the root of both a scaling problem and the whole duplicate-edge bug family.

**The symptom.** The repo's own `Benchmark_Stabilize_nestedBinds` scales badly, and
did so before this pass:

| depth | base | now |
|---:|---:|---:|
| 16 | 263us | 147us |
| 32 | 2.11ms | 1.02ms |
| 64 | 21.7ms | 10.9ms |
| 128 | 293ms | 155ms |

This pass made it ~1.9x faster without changing the shape: each doubling of depth
costs 7x, then 10x, then 14x. Profiling `nestedBinds_64` puts **50% of the time in
`remove`**, reached through `changeParent` -> `becameUnnecessary` -> `unlink`. The
reason is visible in the graph: that benchmark builds depth^2 binds over a single
shared control var, so one node's dependent list grows quadratically --
`maxChildren` measures 256, 1024 and 4096 at depth 16, 32 and 64 -- and every
unlink scans it.

**What `incremental` does.** Each edge carries its position on *both* ends:

    mutable my_parent_index_in_child_at_index : int array
    mutable my_child_index_in_parent_at_index : int array

`remove_parent ~child ~parent ~child_index` therefore never searches. It reads
`parent.my_parent_index_in_child_at_index.(child_index)` to find where the parent
sits in the child's list, moves the last entry into the freed slot, fixes that
entry's reciprocal index, and decrements the count. O(1), no matter how many
dependents a node has.

Two design choices make that work, and they are the actual clue:

- **The two directions have different requirements, and are represented
  differently.** A node's inputs are positional and fixed-arity: `Kind.iteri_children`
  yields `(slot, child)` pairs derived from the node kind, so map2's inputs are
  always slots 0 and 1. Its dependents are an unordered set, so removal is free to
  swap the last entry into the hole. go-incr conflates the two -- both are
  `[]INode` with order-preserving removal -- and pays an ordered O(k) removal for
  the direction that never needed ordering.
- **An edge is identified by (node, slot), not by node identity.** `Map3(x, x, x)`
  is three independent edges at slots 0, 1 and 2, added and removed one at a time.
  There is no "remove every edge to node X" operation, which is precisely the
  operation that produced all three duplicate-edge bugs in this codebase: teardown
  running once per duplicate, a node both queued and recomputed directly, and the
  node counter double-decrementing. In the reference's model those bugs are not
  guarded against -- they are unrepresentable.

**Direction for go-incr.** The ingredients are already present: `IParents.Parents()`
is the ordered, fixed-arity input list, so the slot index is available at every
call site that currently passes a node identity. A migration would be:

1. Give `unlink`/`removeParent`/`removeChild` a slot index instead of an
   `Identifier`, taking it from the position in `Parents()`.
2. Add the two reciprocal index arrays to `Node`, maintained by `link`/`unlink`.
   They can live inline for the common small-arity case, like `parentsInline`.
3. Drop order preservation for `children`, since dependents are unordered.

That makes `unlink` O(1), removes the last identity-keyed search from the hot
path, and retires the duplicate-edge hazard class rather than testing around it.
It touches the whole linking layer, so it wants to be its own change with its own
verification rather than being folded into a performance pass.

### E. Admit bind main nodes to the direct path

The one remaining scheduling difference. `incremental` allows a bind main node to
be recomputed directly when `node.height > b.lhs_change.height`, i.e. its
lhs-change node has already run. go-incr excludes bind mains outright, because
`IBindMain` does not expose the lhs-change height and the ordering constraint is
not visible in the node's inputs. Worth one node per bind — low value, and needs
the interface widened to do safely.

### F. (done) Avoid the redundant isStale() on propagation

`incremental` does not re-derive staleness when propagating to dependents -- it knows they are
stale because their input just changed, and only asserts it in debug builds. go-incr's
`shouldRecomputeChild` called `isStale()`, which walks the node's inputs; on a `map2` tree that
is an input scan per dependent per visit.

What landed: `shouldRecomputeChild` now short-circuits before the scan. It tests heap
membership and necessity first, both single field loads, and then takes the fast path
`cn.staler == nil && cn.recomputedAt < stabilizationNum` -- the caller has just stamped the
input's `changedAt` with the current pass, so a node without a custom staleness rule that has
not run this pass is necessarily stale. The input scan is reached only for nodes that supply
their own `IStale`, which is what the care this section asked for amounts to: the delegate's
answer cannot be inferred, so it still has to be asked.

### G. Kind-directed dispatch

`incremental` recomputes via a match on a `Kind` variant, so the compiler emits a jump table
and the per-kind logic inlines. go-incr dispatches through cached interface values --
`stabilizer`, `cutoffer`, `staler` on `Node` -- which cannot inline. (Those were bound method
values once, which also allocated a closure per node; section B replaced them with interfaces.)
This is idiomatic Go and probably not worth changing, but it is a standing part of the
remaining constant factor, and section N argues the larger remaining term is footprint rather
than dispatch.

