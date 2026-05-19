# ログ出力仕様

## 概要

本プロジェクトでは、用途に応じて以下の2つのログ出力方式を使用しています。

| 方式 | 使用箇所 | 出力先 |
|------|----------|--------|
| Go 標準 `log` パッケージ | サーバー起動・初期化・致命的エラー | stderr（プレフィックス付き） |
| Gin カスタムロガーミドルウェア | HTTPリクエスト/レスポンス | stdout |

---

## 1. 標準 `log` パッケージ

### ライブラリ

```go
import "log"
```

Go 標準ライブラリをそのまま使用しています。サードパーティのロギングライブラリは導入されていません。

### 出力フォーマット

Go 標準の `log` パッケージはデフォルトで以下のプレフィックスを付与します。

```
2009/11/10 23:00:00 <メッセージ>
```

- タイムスタンプ: `YYYY/MM/DD HH:MM:SS`（ローカルタイム）
- フラグ: `log.LstdFlags`（日付 + 時刻）

### 使用関数

| 関数 | 動作 | 使用ケース |
|------|------|-----------|
| `log.Printf(format, ...)` | ログ出力 | 情報・警告メッセージ |
| `log.Fatalf(format, ...)` | ログ出力後に `os.Exit(1)` | 回復不能なエラー |
| `log.Println(...)` | ログ出力（改行付き） | 単純なメッセージ |
| `log.Fatal(...)` | ログ出力後に `os.Exit(1)` | 回復不能なエラー |

### 使用箇所と出力内容

#### `cmd/server/main.go` — サーバー起動

| タイミング | レベル | メッセージ例 |
|-----------|--------|-------------|
| Datastoreサービス初期化失敗 | FATAL | `Failed to initialize Datastore service: <error>` |
| Firebaseサービス初期化失敗 | FATAL | `Failed to initialize Firebase service: <error>` |
| サーバー起動 | INFO | `Cloud Run Kyouen Server starting on port <port>` |
| 環境情報 | INFO | `Environment: <gin_mode>` |
| プロジェクトID | INFO | `Project ID: <project_id>` |
| サーバー起動失敗 | FATAL | `Failed to start server: <error>` |

#### `internal/config/config.go` — 設定読み込み

| タイミング | レベル | メッセージ例 |
|-----------|--------|-------------|
| バリデーション失敗 | FATAL | `Configuration validation failed: <error>` |
| PROJECT_ID未設定 | WARNING | `Warning: GOOGLE_CLOUD_PROJECT is not set, using default: my-android-server` |

#### `internal/tracing/tracing.go` — OpenTelemetryトレーシング初期化

| タイミング | レベル | メッセージ例 |
|-----------|--------|-------------|
| Cloud Trace Exporter作成失敗 | WARNING | `Warning: failed to create Cloud Trace exporter: <error> (tracing disabled)` |
| TracerProviderシャットダウン失敗 | ERROR | `Error shutting down tracer provider: <error>` |

#### `cmd/notify/main.go` — プッシュ通知

| タイミング | レベル | メッセージ例 |
|-----------|--------|-------------|
| 新着ステージなし | INFO | `No new stages registered, skipping notification` |
| 新着ステージ検出 | INFO | `Found <n> new stage(s) <list>, sending push notification to topic: <topic>` |
| 通知送信成功 | INFO | `Push notification sent successfully` |
| 各種初期化/処理失敗 | FATAL | `Failed to ...: <error>` |

---

## 2. Gin HTTPリクエストロガー

### 実装箇所

`internal/middleware/logger.go`

### 登録方法

`cmd/server/main.go` にてミドルウェアとして登録しています。

```go
router.Use(otelgin.Middleware("kyouen-server"))  // OpenTelemetryトレーシング
router.Use(middleware.Logger())                   // カスタムHTTPアクセスログ
router.Use(gin.Recovery())                        // パニックリカバリー
```

### 出力フォーマット

```
<ClientIP> - [<Timestamp>] "<Method> <Path> <Protocol> <StatusCode> <Latency> "<UserAgent>" <ErrorMessage>"
```

#### フィールド詳細

| フィールド | 内容 | 例 |
|-----------|------|-----|
| `ClientIP` | リクエスト元IPアドレス | `192.168.1.1` |
| `Timestamp` | リクエスト処理時刻（RFC1123形式） | `Mon, 19 May 2026 12:00:00 JST` |
| `Method` | HTTPメソッド | `GET`, `POST` |
| `Path` | リクエストパス（クエリ文字列含む） | `/v2/stages?start_stage_no=0` |
| `Protocol` | HTTPプロトコルバージョン | `HTTP/1.1` |
| `StatusCode` | HTTPレスポンスステータスコード | `200`, `400`, `500` |
| `Latency` | レスポンス処理時間 | `1.234ms` |
| `UserAgent` | リクエスト元UserAgent文字列 | `okhttp/4.9.0` |
| `ErrorMessage` | Ginのエラーメッセージ（エラー時のみ） | 通常は空文字列 |

#### 出力例

```
192.168.1.1 - [Mon, 19 May 2026 12:00:00 JST] "GET /v2/stages HTTP/1.1 200 1.234ms "okhttp/4.9.0" "
```

---

## 3. OpenTelemetry トレーシング

`otelgin.Middleware("kyouen-server")` により、全HTTPリクエストのトレースが Google Cloud Trace に送信されます。
これはログとは独立した計装であり、Cloud Console の「トレース」から確認できます。

- サンプリング: `AlwaysSample`（全リクエストをトレース）
- エクスポーター: Google Cloud Trace（Cloud Run 環境）
- Cloud Trace Exporter 初期化失敗時はトレーシング無効で継続（ログに Warning を出力）

---

## 4. GIN_MODEとログ出力の関係

`GIN_MODE` 環境変数（`cfg.Environment`）により Gin の動作モードが変わります。

| `GIN_MODE` | 動作 | Gin内部ログ |
|-----------|------|------------|
| `debug`（デフォルト） | 開発モード | ルート登録情報などのデバッグログを出力 |
| `release` | 本番モード | デバッグログを抑制 |

Cloud Run 本番環境では `GIN_MODE=release` を設定することを推奨します。

---

## 5. エラーハンドリングとログの関係

ハンドラー層（`internal/stage/handler.go` 等）では、エラー発生時に `log` への出力は行わず、
HTTPレスポンスのJSONボディにエラー情報を含めて返します。

```go
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
```

アクセスログミドルウェアがレスポンスステータスコードを記録するため、エラーはアクセスログから確認できます。
