#!/usr/bin/env python3
"""Join the two harnesses' JSONL output into a comparison table.

Usage: compare.py go.jsonl ocaml.jsonl
"""
import json
import sys


def load(path):
    out = {}
    with open(path) as f:
        for line in f:
            line = line.strip()
            if line:
                r = json.loads(line)
                out[r["name"]] = r
    return out


def human(ns):
    if ns < 1_000:
        return f"{ns:.0f}ns"
    if ns < 1_000_000:
        return f"{ns/1_000:.1f}us"
    return f"{ns/1_000_000:.2f}ms"


def main():
    go, oc = load(sys.argv[1]), load(sys.argv[2])
    # Preserve the order the Go harness emitted, then append Go-only cases.
    names = list(go.keys()) + [n for n in oc if n not in go]

    print(f"{'case':<38} {'go-incr':>10} {'ocaml':>10} {'ratio':>8}  winner")
    print("-" * 82)
    for name in names:
        g, o = go.get(name), oc.get(name)
        if g is None or o is None:
            only = "go-incr only" if o is None else "ocaml only"
            have = g or o
            print(f"{name:<38} {human(have['min_ns']):>10} {'-':>10} {'-':>8}  ({only})")
            continue
        gn, on = g["min_ns"], o["min_ns"]
        ratio = gn / on
        if ratio > 1:
            winner = f"ocaml {ratio:.1f}x faster"
        else:
            winner = f"go-incr {1/ratio:.1f}x faster"
        print(f"{name:<38} {human(gn):>10} {human(on):>10} {ratio:>8.2f}  {winner}")


if __name__ == "__main__":
    main()
