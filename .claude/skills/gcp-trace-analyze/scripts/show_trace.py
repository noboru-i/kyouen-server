#!/usr/bin/env python3
"""
Cloud Trace v1 の単一 trace → span ツリーを整形して出力。

Usage:
  curl .../traces/{traceId} | python3 show_trace.py
"""
import sys, json, re
from datetime import datetime


def parse(s):
    # Cloud Trace timestamps may have up to 9 fractional digits (RFC3339Nano);
    # fromisoformat only handles up to 6, so truncate the excess.
    s = re.sub(r'(\.\d{6})\d+', r'\1', s.replace("Z", "+00:00"))
    return datetime.fromisoformat(s)


def main():
    data = json.load(sys.stdin)
    spans = data.get("spans", [])
    if not spans:
        print("No spans found.")
        return

    spans.sort(key=lambda s: parse(s["startTime"]))
    root = next((s for s in spans if not s.get("parentSpanId")), spans[0])
    root_start = parse(root["startTime"])
    total_ms = (max(parse(s["endTime"]) for s in spans) - root_start).total_seconds() * 1000

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
