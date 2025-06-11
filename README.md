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
- **GitHub CLI** (`gh`) - デプロイスクリプト用
- **Google Cloud Projects:**
  - **DEV環境**: `api-project-732262258565`
  - **本番環境**: `my-android-server`

**デプロイに必要なツール:**
- GitHub CLI: `brew install gh` (macOS) / `sudo apt install gh` (Ubuntu)
- Docker: GitHub Actionsで自動実行（ローカル不要）
- Google Cloud SDK: GitHub Actionsで自動設定（ローカル不要）

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
# 前提: GitHub CLIのインストールと認証が必要
# brew install gh && gh auth login

# DEV環境にデプロイ（デフォルト）
./scripts/deploy.sh dev

# 本番環境にデプロイ（確認付き）
./scripts/deploy.sh prod
```

**scripts/deploy.sh の仕組み:**
- GitHub CLI (`gh`) を使用してGitHub Actionsワークフローを実行
- ローカルでのDocker build/pushは不要
- 統一されたCI/CDパイプラインを活用
- デプロイ進捗をリアルタイムで監視可能

#### Cloud Build使用
```bash
gcloud builds submit --config cloudbuild.yaml
```

#### 手動デプロイ例（DEV環境）
```bash
docker build -t gcr.io/api-project-732262258565/kyouen-server:latest .
docker push gcr.io/api-project-732262258565/kyouen-server:latest
gcloud run deploy kyouen-server-dev \
  --image gcr.io/api-project-732262258565/kyouen-server:latest \
  --region asia-northeast1 \
  --allow-unauthenticated \
  --set-env-vars GOOGLE_CLOUD_PROJECT=api-project-732262258565,ENVIRONMENT=dev
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

- **自動デプロイ**
  - **DEV環境** (`.github/workflows/deploy-dev.yml`): mainブランチ自動デプロイ
  - **本番環境** (`.github/workflows/deploy-prod.yml`): 手動実行 + 確認入力必須
  - **共通処理** (`.github/workflows/deploy-common.yml`): 再利用可能ワークフロー
  - Workload Identity認証（環境別）
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

### ローカル開発用設定

`.env.example`を`.env`にコピーしてローカル環境用の設定を行ってください：

```bash
cp .env.example .env
```

`.env`ファイルの設定項目：

```bash
# サーバー設定
PORT=8080                          # サーバーポート
GIN_MODE=debug                     # 開発時はdebug、本番ではrelease

# Google Cloud設定
GOOGLE_CLOUD_PROJECT=my-android-server  # プロジェクトID
FIRESTORE_EMULATOR_HOST=localhost:8080  # エミュレータ使用時

# Twitter OAuth設定
CONSUMER_KEY=your_twitter_consumer_key
CONSUMER_SECRET=your_twitter_consumer_secret

# Firebase設定
FIREBASE_CREDENTIALS_FILE=path/to/service-account.json
```

**注意**: `.env`ファイルは`.gitignore`に含まれているため、機密情報を安全に管理できます。

### Cloud Run用環境変数

#### 必須設定
```bash
# 本番環境
GOOGLE_CLOUD_PROJECT=my-android-server
ENVIRONMENT=prod

# DEV環境
GOOGLE_CLOUD_PROJECT=api-project-732262258565
ENVIRONMENT=dev

# 共通
PORT=8080
```

#### オプション設定
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