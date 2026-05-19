#!/usr/bin/env python3
"""
Cloud Trace v1 の単一 trace → span ツリーを整形して出力。

Usage:
  curl .../traces/{traceId} | python3 show_trace.py
"""
import sys, json
from datetime import datetime


def parse(s):
    return datetime.fromisoformat(s.replace("Z", "+00:00"))


def main():
    data = json.load(sys.stdin)
    spans = data.get("spans", [])
    if not spans:
        print("No spans found.")
        return

    spans.sort(key=lambda s: parse(s["startTime"]))
    root_start = parse(spans[0]["startTime"])
    total_ms = (parse(spans[0]["endTime"]) - root_start).total_seconds() * 1000

    print(f"TRACE: {data.get('traceId','?')}  total={total_ms:.1f}ms  spans={len(spans)}")
    print(f"{'offset':>9} {'dur':>9} {'name':<60} {'parent':<12}")
    print("-" * 95)

    for span in spans:
        off = (parse(span["startTime"]) - root_start).total_seconds() * 1000
        dur = (parse(span["endTime"]) - parse(span["startTime"])).total_seconds() * 1000
        pid = span.get("parentSpanId", "")
        print(f"{off:>7.1f}ms {dur:>7.1f}ms {span.get('name','?')[:60]:<60} {pid[-8:] if pid else 'ROOT':<12}")


if __name__ == "__main__":
    main()
