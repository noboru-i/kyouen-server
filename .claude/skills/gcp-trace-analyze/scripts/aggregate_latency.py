#!/usr/bin/env python3
"""
Cloud Trace v1 ROOTSPAN レスポンス → エンドポイント別 p50/p95/p99/max を出力。

Usage:
  curl ... | python3 aggregate_latency.py
"""
import sys, json, statistics
from collections import defaultdict
from datetime import datetime

SKIP = ["BatchWriteSpans", "BatchWriteS", ".well-known", "/html/"]


def parse(s):
    return datetime.fromisoformat(s.replace("Z", "+00:00"))


def main():
    data = json.load(sys.stdin)
    traces = data.get("traces", [])
    print(f"total traces: {len(traces)}")

    by_name = defaultdict(list)
    for t in traces:
        for span in t.get("spans", []):
            name = span.get("name", "?")
            if any(k in name for k in SKIP):
                continue
            try:
                ms = (parse(span["endTime"]) - parse(span["startTime"])).total_seconds() * 1000
                by_name[name].append(ms)
            except Exception:
                pass

    if not by_name:
        print("No spans found.")
        return

    print()
    hdr = f"{'endpoint':<55} {'count':>6} {'p50':>10} {'p95':>10} {'p99':>10} {'max':>10}"
    print(hdr)
    print("-" * len(hdr))

    for name, durs in sorted(by_name.items(), key=lambda x: -statistics.median(x[1])):
        s = sorted(durs)
        n = len(s)
        print(f"{name[:55]:<55} {n:>6} "
              f"{s[n//2]:>8.1f}ms "
              f"{s[min(n-1,int(n*.95))]:>8.1f}ms "
              f"{s[min(n-1,int(n*.99))]:>8.1f}ms "
              f"{s[-1]:>8.1f}ms")


if __name__ == "__main__":
    main()
