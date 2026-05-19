#!/usr/bin/env python3
"""
ROOTSPAN レスポンス → endpoint ごとの最遅 trace ID を抽出。

Usage:
  curl ... | python3 extract_slow_traces.py
"""
import sys, json
from datetime import datetime

SKIP = ["BatchWriteSpans", "BatchWriteS", ".well-known", "/html/"]


def parse(s):
    return datetime.fromisoformat(s.replace("Z", "+00:00"))


def main():
    data = json.load(sys.stdin)
    rows = []
    for t in data.get("traces", []):
        for span in t.get("spans", []):
            name = span.get("name", "?")
            if any(k in name for k in SKIP):
                continue
            try:
                ms = (parse(span["endTime"]) - parse(span["startTime"])).total_seconds() * 1000
                rows.append((ms, name, t["traceId"]))
            except Exception:
                pass

    rows.sort(reverse=True)
    seen = {}
    for ms, name, tid in rows:
        key = name.split("?")[0]
        if key not in seen:
            seen[key] = (ms, tid)

    for k, (ms, tid) in seen.items():
        print(f"{ms:>7.1f}ms  {k:<50}  {tid}")


if __name__ == "__main__":
    main()
