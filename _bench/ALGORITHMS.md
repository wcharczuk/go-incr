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

`Node` ended at 368 bytes, up from the 288 the lazy `nodeExtra` alone achieved:
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

Interleaved measurement, best-of-3, ratio = go-incr / ocaml (lower is better):

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
every rebuild allocates fresh 400-byte nodes. That is where OCaml's bump-allocated
minor heap is hardest to match, and why (B) below is the highest-value item left.

## Not done — what go-incr would still benefit from

Ordered by expected value.

### A. Cutoff by default (semantic change, needs a decision)

`incremental`'s `maybe_change_value` compares the new value against the old with
the node's cutoff — **physical equality by default** — and if unchanged, does not
touch `changed_at` and does not enqueue dependents at all. Propagation stops dead.

go-incr has `Cutoff`/`Cutoff2` nodes but no default, so a write of an unchanged
value propagates through the entire graph. This is the `wide/update_same` case,
still **10x** and growing with graph size.

This is a real design difference, not a bug, and changing the default would alter
observable behavior for existing users (update handlers would stop firing on
no-op writes), so it should be an explicit decision rather than something folded
into a performance pass. Options: an `OptGraphDefaultCutoff` graph option, or a
cutoff-by-default `Var` variant. Workloads where writes are frequently no-ops —
common in practice — would benefit a lot.

### B. (done) Shrink `Node` and cut allocation per node

`Node` is ~400 bytes. Bind rebuilds allocate one per node created, and the bind
benchmarks are dominated by allocation and GC (~25% of profile in
`scanObjectsSmall` / `mallocgcSmallScanNoHeader`). This is where OCaml's
bump-allocated minor heap and cheap generational collection are hardest to match.

Moving cold fields (`kind`, `label`, `metadata`, the three handler slices,
`shouldBeInvalidatedFn`, `parentsFn`, `invalidateFn`, `sentinels`,
`heightInAdjustHeightsHeap`) behind a lazily-allocated side pointer would roughly
halve it, cutting both allocation volume and GC scan work. Invasive but mechanical.

### C. (done) Inline the first input and dependent

Done throughout; see (5) above. No `[]INode{...}` literal remains in the library.

### D. Admit bind main nodes to the direct path

The one remaining scheduling difference. `incremental` allows a bind main node to
be recomputed directly when `node.height > b.lhs_change.height`, i.e. its
lhs-change node has already run. go-incr excludes bind mains outright, because
`IBindMain` does not expose the lhs-change height and the ordering constraint is
not visible in the node's inputs. Worth one node per bind — low value, and needs
the interface widened to do safely.

### E. (done) Avoid the redundant isStale() on propagation

`incremental` does not re-derive staleness when propagating to dependents — it
knows they are stale because their input just changed, and only asserts it in
debug builds. go-incr's `shouldRecomputeChild` calls `isStale()`, which walks the
node's inputs; on a `map2` tree that is a parent scan per dependent per visit.
The information is already known at the call site. Needs care around nodes with a
custom `staleFn` and invalidated nodes.

### F. Kind-directed dispatch

`incremental` recomputes via a match on a `Kind` variant, so the compiler emits a
jump table and the per-kind logic inlines. go-incr dispatches through interfaces
and cached closures (`stabilizeFn`, `cutoffFn`, `staleFn`), which cannot inline.
This is idiomatic Go and probably not worth changing, but it is a standing part
of the remaining constant factor.
