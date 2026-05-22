#!/usr/bin/env python3
"""
ROOTSPAN レスポンス → endpoint ごとの最遅 trace ID を抽出。

Usage:
  curl ... | python3 extract_slow_traces.py
"""
import sys, json, re
from datetime import datetime

SKIP = ["BatchWriteSpans", "BatchWriteS", ".well-known", "/html/"]


def parse(s):
    # Cloud Trace timestamps may have up to 9 fractional digits (RFC3339Nano);
    # fromisoformat only handles up to 6, so truncate the excess.
    s = re.sub(r'(\.\d{6})\d+', r'\1', s.replace("Z", "+00:00"))
    return datetime.fromisoformat(s)


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
