#!/usr/bin/env python3
"""Compare two runs of the same harness to show the effect of a change.

Usage: speedup.py before.jsonl after.jsonl [ocaml.jsonl]

With a third argument, also shows the remaining gap to the OCaml baseline.
"""
import json
import sys


def load(path):
    out = {}
    with open(path) as f:
        for line in f:
            if line.strip():
                r = json.loads(line)
                out[r["name"]] = r["min_ns"]
    return out


def human(ns):
    if ns < 1_000:
        return f"{ns:.0f}ns"
    if ns < 1_000_000:
        return f"{ns/1_000:.1f}us"
    return f"{ns/1_000_000:.2f}ms"


def main():
    before, after = load(sys.argv[1]), load(sys.argv[2])
    ocaml = load(sys.argv[3]) if len(sys.argv) > 3 else {}

    head = f"{'case':<38} {'before':>9} {'after':>9} {'change':>9}"
    if ocaml:
        head += f" {'vs ocaml':>9}"
    print(head)
    print("-" * (len(head) + 2))

    for name, b in before.items():
        if name not in after:
            continue
        a = after[name]
        # Positive percent means the change made this case faster.
        pct = (b - a) / b * 100
        row = f"{name:<38} {human(b):>9} {human(a):>9} {pct:>8.1f}%"
        if name in ocaml:
            row += f" {a/ocaml[name]:>8.2f}x"
        print(row)


if __name__ == "__main__":
    main()
