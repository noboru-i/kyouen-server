# 共円パズルゲーム API サーバー

「共円」パズルゲーム用のREST APIサーバーです。プレイヤーはグリッド上に石を配置し、ちょうど4つの石で円や直線を形成する知的パズルゲームをお楽しみいただけます。

## 🏗️ アーキテクチャ

- **プラットフォーム**: Cloud Run (コンテナベース)
- **フレームワーク**: Gin (Go製高速Webフレームワーク)
- **データベース**: DatastoreモードFirestore (既存データと互換性を保持)
- **言語**: Go 1.23+
- **認証**: Twitter OAuth + Firebase Authentication

## 🚀 クイックスタート

### 前提条件

- Go 1.23以上
- Docker (Cloud Runデプロイ用)
- Google Cloud SDK
- プロジェクトID: `my-android-server`

### ローカル開発

#### 1. デモサーバー（認証不要）
```bash
go run cmd/demo_server/main.go
```
サンプルデータでAPI動作を確認できます。

#### 2. 本番接続サーバー
```bash
go run cmd/server/main.go
```
実際のDatastoreに接続して動作します。

#### 3. テスト用サーバー
```bash
go run cmd/test_server/main.go
```
Datastore接続テスト用です。

**アクセス先:**
- サーバー: http://localhost:8080/

## 🔄 API エンドポイント

### ヘルスチェック
```
GET /health
```

### 統計情報
```
GET /v2/statics
```

### ステージ管理
```
GET /v2/stages              # ステージ一覧取得
POST /v2/stages             # 新規ステージ作成
POST /v2/stages/{id}/clear  # ステージクリア
```

### ユーザー管理
```
POST /v2/users/login        # ログイン
```

## 🎮 ゲームロジック

### 共円判定アルゴリズム
`models/kyouen.go`に実装された核となるアルゴリズム：
- 4つの石が同一直線上にあるかの判定
- 4つの石が同一円周上にあるかの判定
- 回転・反転を考慮した重複ステージの検出

### ステージ表現
```
"0": 空のセル
"1": 黒石（パズルの石）
"2": 白石（ユーザーの解答）
```

## 🚢 デプロイメント

### Cloud Runへのデプロイ

#### 自動デプロイ（推奨）
```bash
./scripts/deploy.sh
```

#### Cloud Build使用
```bash
gcloud builds submit --config cloudbuild.yaml
```

#### 手動デプロイ
```bash
docker build -t gcr.io/my-android-server/kyouen-server:latest .
docker push gcr.io/my-android-server/kyouen-server:latest
gcloud run deploy kyouen-server \
  --image gcr.io/my-android-server/kyouen-server:latest \
  --region asia-northeast1 \
  --allow-unauthenticated
```


## 🧪 テスト

### 全テスト実行
```bash
go test -v ./...
```

### 特定パッケージのテスト
```bash
go test -v ./models
```

### ビルドテスト
```bash
go build -v ./...
```

### カバレッジテスト
```bash
go test -race -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -html=coverage.out
```

## 🚀 CI/CD

### GitHub Actions
プロジェクトはGitHub Actionsによる自動CI/CDを設定済みです：

- **PR検証** (`.github/workflows/pr_validation.yml`)
  - Go 1.23での自動テスト・ビルド
  - 全エントリーポイントのビルド確認
  - Dockerイメージビルドテスト

- **自動デプロイ** (`.github/workflows/deploy.yml`)
  - mainブランチプッシュ時の自動Cloud Runデプロイ
  - Workload Identity認証
  - デプロイ後のヘルスチェック

## 📚 OpenAPI (Swagger)

### Swagger UI表示
```bash
docker run -p 10000:8080 \
  -v $(pwd)/docs:/usr/share/nginx/html/docs \
  -e API_URL=http://localhost:10000/docs/specs/index.yaml \
  swaggerapi/swagger-ui
```

### コード生成

#### Go用構造体生成
```bash
openapi-generator generate -i docs/specs/index.yaml -g go-server -o ./tmp
cp tmp/go/model_*.go openapi
rm -rf tmp
```

#### Androidクライアント用生成
```bash
openapi-generator generate -i docs/specs/index.yaml -g kotlin -o ./tmp \
  --additional-properties="packageName=hm.orz.chaos114.android.tumekyouen.network"
cp -r tmp/src/main/kotlin/hm/orz/chaos114/android/tumekyouen/network/models \
  ../kyouen-android/app/src/main/java/hm/orz/chaos114/android/tumekyouen/network
rm -rf tmp
```

## 📁 プロジェクト構成

```
kyouen-server/
├── cmd/                    # エントリーポイント
│   ├── server/            # Cloud Run本番用
│   ├── demo_server/       # デモ用
│   └── test_server/       # テスト用
├── handlers/              # APIハンドラー
│   └── v2/               # Gin対応API
├── models/               # ゲームロジック
│   └── kyouen.go        # 共円判定アルゴリズム
├── services/            # サービス層
│   └── datastore.go     # Datastoreサービス
├── config/              # 設定管理
├── middleware/          # ミドルウェア
├── openapi/            # OpenAPI生成ファイル
├── docs/               # ドキュメント
├── scripts/            # デプロイスクリプト
└── tasks/              # 移行記録
```

## ⚙️ 環境変数

### 必須設定
```bash
GOOGLE_CLOUD_PROJECT=my-android-server
PORT=8080
```

### オプション設定
```bash
GIN_MODE=release                    # 本番環境用
CONSUMER_KEY=your_twitter_key       # Twitter OAuth
CONSUMER_SECRET=your_twitter_secret # Twitter OAuth
```

## 🔄 移行履歴

このプロジェクトは以下の技術で構築されています：
- **プラットフォーム**: Cloud Run (コンテナベース)
- **フレームワーク**: Gin (Go製高速Webフレームワーク)  
- **データベース**: DatastoreモードFirestore

詳細は以下のドキュメントを参照：
- `tasks/migration-plan.md` - 完了した移行計画
- `tasks/datastore-mode-migration.md` - 移行戦略の変更記録

## 🤝 開発について

詳細な開発ガイドは `CLAUDE.md` を参照してください。