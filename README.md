# 共円パズルゲーム API サーバー

「共円」パズルゲーム用のREST APIサーバーです。プレイヤーはグリッド上に石を配置し、ちょうど4つの石で円や直線を形成する知的パズルゲームをお楽しみいただけます。

> **開発者向け詳細ガイド**: [CLAUDE.md](./CLAUDE.md) を参照してください

## 🏗️ アーキテクチャ

- **プラットフォーム**: Cloud Run (コンテナベース)
- **フレームワーク**: Gin (Go製高速Webフレームワーク)
- **データベース**: DatastoreモードFirestore (既存データと互換性を保持)
- **言語**: Go 1.24+
- **認証**: Twitter OAuth + Firebase Authentication

## 🚀 クイックスタート

### 前提条件

- Go 1.24以上
- Firebase CLI (`firebase`) - ローカルエミュレーター用
- GitHub CLI (`gh`) - デプロイスクリプト用

### ローカル開発

```bash
# エミュレーターを使ったローカル開発（推奨）
gcloud emulators firestore start --database-mode=datastore-mode --host-port=0.0.0.0:9098
firebase emulators:start

DATASTORE_EMULATOR_HOST=localhost:9098 FIREBASE_AUTH_EMULATOR_HOST=localhost:9099 go run cmd/server/main.go

# ローカルアクセス先: http://localhost:8080/
```

### Firebase プロジェクト切り替え

```bash
firebase use dev   # DEV環境 (api-project-732262258565)
firebase use prod  # 本番環境 (my-android-server)
```

## 🔄 API エンドポイント

### ヘルスチェック
```
GET /health
```

### API ドキュメント
```
GET /static/swagger-ui.html   # Swagger UI
GET /docs/specs/index.yaml    # OpenAPI仕様
```

### 統計情報
```
GET /v2/statics
```

### ステージ管理
```
GET  /v2/stages                    # ステージ一覧取得
POST /v2/stages                    # 新規ステージ作成（要認証）
POST /v2/stages/sync               # ステージ同期（要認証）
PUT  /v2/stages/{stageNo}/clear    # ステージクリア（認証任意）
GET  /v2/recent_stages             # 最近のステージ一覧
GET  /v2/activities                # アクティビティ一覧
```

### ユーザー管理
```
POST   /v2/users/login          # ログイン
DELETE /v2/users/delete-account # アカウント削除（要認証）
```

## 🧪 テスト

```bash
# 全テスト実行
go test -v ./...

# ビルドテスト
go build -v ./...
```

## 🏛️ インフラ管理

GCP/Firebase リソースは Terraform で管理しています。

```bash
cd terraform/envs/dev
terraform plan -var-file="terraform.tfvars"
terraform apply -var-file="terraform.tfvars"
```

詳細は [terraform/README.md](./terraform/README.md) を参照してください。

## 🚀 CI/CD

GitHub Actionsによる自動CI/CDを設定済み：
- **PR検証**: Go 1.24での自動テスト・ビルド
- **自動デプロイ**: DEV環境（mainブランチ）、本番環境（手動実行）
- **インフラ**: Terraform で Firebase Auth・Cloud Run・Artifact Registry を管理


## 🤝 開発について

詳細な開発ガイドは [CLAUDE.md](./CLAUDE.md) を参照してください。
