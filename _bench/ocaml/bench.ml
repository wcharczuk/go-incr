(* Cross-library comparison benchmark suite for Jane Street's incremental.

   This is the counterpart to ../go/main.go: it implements the identical set of
   graph shapes and update patterns and emits the same JSON records, so the two
   outputs can be joined on (library, name).

   The state-passing interface (rather than the [Make ()] functor) is used
   throughout so that construction benchmarks can allocate a fresh graph per
   iteration. Each case body keeps its state witness type local, which is why
   every case is a self-contained closure returning [unit -> unit]. *)

module I = Incremental

let min_batch_seconds = 0.01
let min_total_seconds = 0.5
let min_rounds = 5

(* [time_batch op batch] runs [op] [batch] times and returns elapsed seconds. *)
let time_batch op batch =
  let t0 = Unix.gettimeofday () in
  for _ = 1 to batch do
    op ()
  done;
  Unix.gettimeofday () -. t0

(* Grow the batch size until a single timed batch is long enough that
   gettimeofday's ~microsecond resolution is not a meaningful error term. *)
let calibrate op =
  let batch = ref 1 in
  while time_batch op !batch < min_batch_seconds && !batch < 1_000_000_000 do
    batch := !batch * 2
  done;
  !batch

type case =
  { name : string
  ; group : string
  ; size : int
  ; setup : unit -> unit -> unit
  }

let emit ~name ~group ~size ~iters ~ns_per_op ~min_ns =
  Printf.printf
    "{\"library\":\"ocaml-incremental\",\"name\":%S,\"group\":%S,\"size\":%d,\"iters\":%d,\"ns_per_op\":%.4f,\"min_ns\":%.4f}\n%!"
    name
    group
    size
    iters
    ns_per_op
    min_ns

let run case =
  let op = case.setup () in
  (* Warm up before calibrating so the batch size reflects steady state. *)
  for _ = 1 to 3 do
    op ()
  done;
  let batch = calibrate op in
  let rounds = ref 0 in
  let total = ref 0.0 in
  let best = ref infinity in
  let continue_ = ref true in
  while !continue_ do
    let elapsed = time_batch op batch in
    total := !total +. elapsed;
    let per_op = elapsed /. float_of_int batch in
    if per_op < !best then best := per_op;
    incr rounds;
    if !rounds >= min_rounds && !total >= min_total_seconds then continue_ := false
  done;
  let iters = !rounds * batch in
  emit
    ~name:case.name
    ~group:case.group
    ~size:case.size
    ~iters
    ~ns_per_op:(!total /. float_of_int iters *. 1e9)
    ~min_ns:(!best *. 1e9)

(* Observers are kept in this list so the GC never finalizes an observation
   mid-benchmark; [~should_finalize:false] covers the same hazard but retaining
   the values also keeps the surrounding graph reachable. *)
let keep_alive : Obj.t list ref = ref []
let retain (x : 'a) = keep_alive := Obj.repr x :: !keep_alive

(* [tree_reduce nodes] reduces pairwise with map2 until one root remains,
   giving a balanced tree of height log2(n) -- so a single leaf change dirties
   only a logarithmic number of nodes. Mirrors buildTreeReduce in the Go harness. *)
let rec tree_reduce (nodes : ('a, 'w) I.t array) : ('a, 'w) I.t =
  let len = Array.length nodes in
  if len = 1
  then nodes.(0)
  else begin
    let out =
      Array.init ((len + 1) / 2) (fun k ->
        if (2 * k) + 1 < len
        then I.map2 nodes.(2 * k) nodes.((2 * k) + 1) ~f:( + )
        else nodes.(2 * k))
    in
    tree_reduce out
  end

let rec chain node d = if d = 0 then node else chain (I.map node ~f:(fun x -> x + 1)) (d - 1)

(* --- Wide graphs --------------------------------------------------------- *)

(* N leaf vars, each mapped, reduced by a balanced map2 tree. *)
let build_wide state n =
  let vars = Array.init n (fun i -> I.Var.create state i) in
  let leaves = Array.map (fun v -> I.map (I.Var.watch v) ~f:(fun x -> x + 1)) vars in
  let root = tree_reduce leaves in
  let o = I.observe ~should_finalize:false root in
  retain o;
  vars, o

let wide_construct n =
  { name = Printf.sprintf "wide/construct/%d" n
  ; group = "wide"
  ; size = n
  ; setup =
      (fun () () ->
        let module S = (val I.State.create ~max_height_allowed:1024 ()) in
        let _vars, _o = build_wide S.t n in
        I.stabilize S.t)
  }

let wide_update_one n =
  { name = Printf.sprintf "wide/update_one/%d" n
  ; group = "wide"
  ; size = n
  ; setup =
      (fun () ->
        let module S = (val I.State.create ~max_height_allowed:1024 ()) in
        let vars, _o = build_wide S.t n in
        I.stabilize S.t;
        let i = ref 0 in
        fun () ->
          incr i;
          I.Var.set vars.(!i mod n) !i;
          I.stabilize S.t)
  }

let wide_update_all n =
  { name = Printf.sprintf "wide/update_all/%d" n
  ; group = "wide"
  ; size = n
  ; setup =
      (fun () ->
        let module S = (val I.State.create ~max_height_allowed:1024 ()) in
        let vars, _o = build_wide S.t n in
        I.stabilize S.t;
        let i = ref 0 in
        fun () ->
          incr i;
          for j = 0 to n - 1 do
            I.Var.set vars.(j) (!i + j)
          done;
          I.stabilize S.t)
  }

(* Set a leaf to the value it already holds. incremental's default cutoff is
   physical equality, so this should short-circuit entirely; go-incr has no
   default cutoff and will propagate. *)
let wide_update_same n =
  { name = Printf.sprintf "wide/update_same/%d" n
  ; group = "wide_cutoff"
  ; size = n
  ; setup =
      (fun () ->
        let module S = (val I.State.create ~max_height_allowed:1024 ()) in
        let vars, _o = build_wide S.t n in
        I.stabilize S.t;
        fun () ->
          I.Var.set vars.(0) 0;
          I.stabilize S.t)
  }

(* --- Deep graphs --------------------------------------------------------- *)

let deep_construct d =
  { name = Printf.sprintf "deep/construct/%d" d
  ; group = "deep"
  ; size = d
  ; setup =
      (fun () () ->
        let module S = (val I.State.create ~max_height_allowed:(d + 64) ()) in
        let v = I.Var.create S.t 0 in
        let o = I.observe ~should_finalize:false (chain (I.Var.watch v) d) in
        retain o;
        I.stabilize S.t)
  }

let deep_update_one d =
  { name = Printf.sprintf "deep/update_one/%d" d
  ; group = "deep"
  ; size = d
  ; setup =
      (fun () ->
        let module S = (val I.State.create ~max_height_allowed:(d + 64) ()) in
        let v = I.Var.create S.t 0 in
        let o = I.observe ~should_finalize:false (chain (I.Var.watch v) d) in
        retain o;
        I.stabilize S.t;
        let i = ref 0 in
        fun () ->
          incr i;
          I.Var.set v !i;
          I.stabilize S.t)
  }

(* --- Bind graphs --------------------------------------------------------- *)

(* Toggling the bind's left-hand side tears down the previous subgraph and
   builds a fresh chain of [d] nodes. *)
let bind_swap_chain d =
  { name = Printf.sprintf "bind/swap_chain/%d" d
  ; group = "bind"
  ; size = d
  ; setup =
      (fun () ->
        let module S = (val I.State.create ~max_height_allowed:(d + 128) ()) in
        let st = S.t in
        let sel = I.Var.create st 0 in
        let b = I.bind (I.Var.watch sel) ~f:(fun which -> chain (I.return st which) d) in
        let o = I.observe ~should_finalize:false b in
        retain o;
        I.stabilize st;
        let i = ref 0 in
        fun () ->
          incr i;
          I.Var.set sel !i;
          I.stabilize st)
  }

(* Many independent binds over one shared left-hand side, summed by a tree, so a
   single set rebuilds n subgraphs at once. *)
let bind_wide_build st n =
  let sel = I.Var.create st 0 in
  let binds =
    Array.init n (fun j ->
      I.bind (I.Var.watch sel) ~f:(fun which -> I.map (I.return st (which + j)) ~f:(fun x -> x + 1)))
  in
  let o = I.observe ~should_finalize:false (tree_reduce binds) in
  retain o;
  sel, o

let bind_wide_swap n =
  { name = Printf.sprintf "bind/wide_swap/%d" n
  ; group = "bind_wide"
  ; size = n
  ; setup =
      (fun () ->
        let module S = (val I.State.create ~max_height_allowed:1024 ()) in
        let sel, _o = bind_wide_build S.t n in
        I.stabilize S.t;
        let i = ref 0 in
        fun () ->
          incr i;
          I.Var.set sel !i;
          I.stabilize S.t)
  }

let bind_wide_construct n =
  { name = Printf.sprintf "bind/wide_construct/%d" n
  ; group = "bind_wide"
  ; size = n
  ; setup =
      (fun () () ->
        let module S = (val I.State.create ~max_height_allowed:1024 ()) in
        let _sel, _o = bind_wide_build S.t n in
        I.stabilize S.t)
  }

let cases =
  List.concat
    [ List.concat_map
        (fun n -> [ wide_construct n; wide_update_one n; wide_update_all n; wide_update_same n ])
        [ 1024; 16384 ]
      (* Depth 1 isolates the fixed per-stabilization overhead from the marginal
         per-node recompute cost that the deeper sizes measure. *)
    ; List.concat_map (fun d -> [ deep_construct d; deep_update_one d ]) [ 1; 128; 2048 ]
    ; List.map bind_swap_chain [ 64; 512 ]
    ; List.map bind_wide_swap [ 256; 4096 ]
    ; [ bind_wide_construct 4096 ]
    ]

(* Prints the observed value of each graph shape after a known mutation. The Go
   harness prints the same lines under -verify; if the two agree we know both
   libraries are building equivalent graphs and actually propagating, rather
   than one of them quietly stabilizing a graph with no necessary nodes. *)
let verify () =
  let module S = (val I.State.create ~max_height_allowed:1024 ()) in
  let vars, o = build_wide S.t 1024 in
  I.stabilize S.t;
  Printf.printf "wide/1024 initial=%d\n" (I.Observer.value_exn o);
  I.Var.set vars.(5) 999;
  I.stabilize S.t;
  Printf.printf "wide/1024 after_set=%d\n" (I.Observer.value_exn o);
  let module D = (val I.State.create ~max_height_allowed:256 ()) in
  let dv = I.Var.create D.t 0 in
  let dobs = I.observe ~should_finalize:false (chain (I.Var.watch dv) 128) in
  retain dobs;
  I.stabilize D.t;
  Printf.printf "deep/128 initial=%d\n" (I.Observer.value_exn dobs);
  I.Var.set dv 7;
  I.stabilize D.t;
  Printf.printf "deep/128 after_set=%d\n" (I.Observer.value_exn dobs);
  let module B = (val I.State.create ~max_height_allowed:256 ()) in
  let st = B.t in
  let sel = I.Var.create st 0 in
  let bobs =
    I.observe
      ~should_finalize:false
      (I.bind (I.Var.watch sel) ~f:(fun which -> chain (I.return st which) 64))
  in
  retain bobs;
  I.stabilize st;
  Printf.printf "bind/64 initial=%d\n" (I.Observer.value_exn bobs);
  I.Var.set sel 3;
  I.stabilize st;
  Printf.printf "bind/64 after_set=%d\n" (I.Observer.value_exn bobs);
  let module W = (val I.State.create ~max_height_allowed:1024 ()) in
  let sel2, wobs = bind_wide_build W.t 256 in
  I.stabilize W.t;
  Printf.printf "bind_wide/256 initial=%d\n" (I.Observer.value_exn wobs);
  I.Var.set sel2 3;
  I.stabilize W.t;
  Printf.printf "bind_wide/256 after_set=%d\n" (I.Observer.value_exn wobs)

(* Reports, for each graph shape, how many nodes the initial stabilization
   recomputes and how many a single subsequent mutation recomputes. The Go
   harness prints the same lines under -stats; matching counts mean the two
   libraries perform equivalent work per stabilization.

   incremental splits its direct-recompute counter in two (one-child and
   min-height); they are summed here because go-incr reports a single total. *)
let stats () =
  let report name state mutate =
    let recomputed () = I.State.num_nodes_recomputed state in
    let direct () =
      I.State.num_nodes_recomputed_directly_because_one_child state
      + I.State.num_nodes_recomputed_directly_because_min_height state
    in
    I.stabilize state;
    let initial = recomputed () in
    let initial_direct = direct () in
    mutate ();
    I.stabilize state;
    Printf.printf
      "%s initial_recomputed=%d update_recomputed=%d update_direct=%d total_nodes=%d\n"
      name
      initial
      (recomputed () - initial)
      (direct () - initial_direct)
      (I.State.num_nodes_created state)
  in
  let module W = (val I.State.create ~max_height_allowed:1024 ()) in
  let vars, _o = build_wide W.t 1024 in
  report "wide/1024" W.t (fun () -> I.Var.set vars.(5) 999);
  List.iter
    (fun d ->
      let module D = (val I.State.create ~max_height_allowed:(d + 64) ()) in
      let v = I.Var.create D.t 0 in
      let o = I.observe ~should_finalize:false (chain (I.Var.watch v) d) in
      retain o;
      report (Printf.sprintf "deep/%d" d) D.t (fun () -> I.Var.set v 7))
    [ 128; 2048 ];
  let module B = (val I.State.create ~max_height_allowed:256 ()) in
  let st = B.t in
  let sel = I.Var.create st 0 in
  let bo =
    I.observe
      ~should_finalize:false
      (I.bind (I.Var.watch sel) ~f:(fun which -> chain (I.return st which) 64))
  in
  retain bo;
  report "bind/swap_chain/64" st (fun () -> I.Var.set sel 3);
  let module V = (val I.State.create ~max_height_allowed:1024 ()) in
  let sel2, _wo = bind_wide_build V.t 256 in
  report "bind/wide_swap/256" V.t (fun () -> I.Var.set sel2 3);
  let module S = (val I.State.create ~max_height_allowed:1024 ()) in
  let vars2, _o2 = build_wide S.t 1024 in
  report "wide/1024/update_same" S.t (fun () -> I.Var.set vars2.(0) 0)

let () =
  if Array.length Sys.argv > 1 && Sys.argv.(1) = "-verify"
  then verify ()
  else if Array.length Sys.argv > 1 && Sys.argv.(1) = "-stats"
  then stats ()
  else
  let filter = if Array.length Sys.argv > 1 then Some Sys.argv.(1) else None in
  let matches name =
    match filter with
    | None -> true
    | Some f ->
      let nl = String.length name and fl = String.length f in
      let rec go i = i + fl <= nl && (String.sub name i fl = f || go (i + 1)) in
      go 0
  in
  List.iter (fun c -> if matches c.name then run c) cases
