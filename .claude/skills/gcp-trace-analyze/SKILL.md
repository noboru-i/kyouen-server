---
name: gcp-trace-analyze
description: >
  Google Cloud Trace REST API v1 を使い、Cloud Run サーバーのエンドポイント別 latency
  (p50/p95/p99/max) を集計し、遅い trace の span 階層から直列 Datastore 呼び出し・N+1・
  Cold Start 等のボトルネックを特定して具体的な改善案を出す。
  「Cloud Trace を確認して」「GCP のトレースでボトルネックを特定して」「Cloud Run の
  レイテンシを調査して」「/gcp-trace-analyze」と言われたら必ずこのスキルを使うこと。
  `gcloud trace traces list` は存在しないため REST API を curl で直接叩く。
allowed-tools: Bash
---

# GCP Cloud Trace 分析スキル

3 フェーズで進む: **集計 → 内訳確認 → 改善案**

---

## フェーズ 0: 準備

**プロジェクト ID**: CLAUDE.md の DEV: `api-project-732262258565` / PROD: `my-android-server` を参照。
不明な場合は `gcloud config get-value project` で取得してユーザーに確認する。

**認証**:
```bash
TOKEN=$(gcloud auth print-access-token)
```
失敗した場合はユーザーに `! gcloud auth login` を依頼する。

**日付範囲** (GNU date を使う。macOS BSD date は `-v` フラグ):
```bash
START=$(date -u -d '7 days ago' +"%Y-%m-%dT%H:%M:%SZ")
END=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
```
GNU date が無い場合:
```bash
START=$(python3 -c "from datetime import datetime,timedelta,timezone; print((datetime.now(timezone.utc)-timedelta(days=7)).strftime('%Y-%m-%dT%H:%M:%SZ'))")
```

---

## フェーズ 1: エンドポイント別 latency 集計

`gcloud trace traces list` は存在しない。必ず curl で REST API を直接叩くこと。

```bash
curl -sG "https://cloudtrace.googleapis.com/v1/projects/${PROJECT}/traces" \
  --data-urlencode "startTime=$START" \
  --data-urlencode "endTime=$END" \
  --data-urlencode "view=ROOTSPAN" \
  --data-urlencode "pageSize=500" \
  --data-urlencode "orderBy=start desc" \
  -H "Authorization: Bearer $TOKEN" \
  | python3 "${CLAUDE_SKILL_DIR}/scripts/aggregate_latency.py"
```

- `BatchWriteSpans` (Cloud Trace 自身の RPC) は自動除外
- 件数が少なければ窓を 14 日に広げる
- root span が二重（`/v2/foo` と `GET /v2/foo`）なら App Engine → Cloud Run の二段計装

---

## フェーズ 2: 遅い trace の内訳確認

### endpoint ごとの最遅 trace ID 抽出

```bash
curl -sG "https://cloudtrace.googleapis.com/v1/projects/${PROJECT}/traces" \
  --data-urlencode "startTime=$START" \
  --data-urlencode "endTime=$END" \
  --data-urlencode "view=ROOTSPAN" \
  --data-urlencode "pageSize=500" \
  -H "Authorization: Bearer $TOKEN" \
  | python3 "${CLAUDE_SKILL_DIR}/scripts/extract_slow_traces.py"
```

### 個別 trace の span 階層取得

```bash
curl -sG "https://cloudtrace.googleapis.com/v1/projects/${PROJECT}/traces/${TRACE_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  | python3 "${CLAUDE_SKILL_DIR}/scripts/show_trace.py"
```

### span 構造の読み方

出力: `offset  duration  span_name  parent_id_suffix`

- **同じ offset** で始まる span → **並列実行**
- **前の span 終了直後に次の offset** → **直列実行**

```
  0.0ms  2128ms  PUT /v2/stages/:stageNo/clear   ROOT
237.7ms   177ms  datastore.Query.GetAll           02330721  ← 直列
414.3ms   737ms  datastore.datastoreClient.Lookup  02330721  ← 直列 (GetUserByID)
1151ms    204ms  datastore.Query.GetAll            02330721  ← 直列
1354ms    774ms  datastore.PutMulti               02330721  ← 直列
```

---

## フェーズ 3: 改善案の提示

以下のパターンを判定して、ファイル:行番号 付きで改善案を提示する。

### パターン A: 独立した Datastore 呼び出しの直列化

**判定**: 互いの結果を使わない複数の Datastore span が直列

**改善**: `errgroup.WithContext(ctx)` で並列化

```go
var resultA TypeA
var resultB TypeB
g, gctx := errgroup.WithContext(ctx)
g.Go(func() error {
    var err error
    resultA, err = service.GetFoo(gctx, ...)
    return err
})
g.Go(func() error {
    var err error
    resultB, err = service.GetBar(gctx, ...)
    return err
})
if err := g.Wait(); err != nil {
    return nil, err
}
```

片方がエラーになると `gctx` がキャンセルされ、他方の Datastore 呼び出しも止まる。

### パターン B: ループ内 Datastore 呼び出し (N+1)

**判定**: 同種の Datastore span が数十本直列に並ぶ

**改善**: キーを先に収集して `GetMulti` / `IN` クエリで一括取得。書き込みは `PutMulti`。

### パターン C: Cold Start による前段オーバーヘッド

**判定**: HTTP span 開始から最初の Datastore span まで 200ms 以上、かつ他の trace では短い

**改善**: Cloud Run `--min-instances=1`（コスト要確認）、または main() での Datastore 事前 warm-up。

### パターン D: 低更新頻度データの毎回読み取り

**判定**: 同じエンティティを毎リクエスト読んでいる（設定値・集計カウンタ等）

**改善**: `sync.Mutex` + TTL 付きメモリキャッシュ。HTTP `Cache-Control: public, max-age=N`。

### パターン E: Cloud Trace 自己計装ノイズ

**判定**: `BatchWriteSpans` がトレース全体の 50%+ を占める

**改善**: `AlwaysSample` → `ParentBased(TraceIDRatioBased(0.1))` でサンプリング率を下げる。

---

## 出力フォーマット

1. **エンドポイント別 latency 表** (p50/p95/p99/max)
2. **遅い trace の内訳** (上位 3〜5 endpoint を show_trace.py 形式で)
3. **改善案リスト** (パターン名・ファイル:行番号・コードスニペット・想定削減 ms)

---

## トラブルシューティング

| 症状 | 対処 |
|---|---|
| `total traces: 0` | 期間を広げる。Cloud Trace 導入直後は件数が少ない場合がある |
| `gcloud auth` 失敗 | `! gcloud auth login` または `! gcloud auth application-default login` |
| `date: invalid option` | Python で代替（フェーズ 0 参照） |
| 403 エラー | `cloudtrace.viewer` ロールを確認 |
| root span 二重 | App Engine → Cloud Run 経由。内側の otelgin span を参照 |
