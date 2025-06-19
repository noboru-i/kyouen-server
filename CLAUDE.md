# CLAUDE.md

このファイルは、このリポジトリでコードを扱う際のClaude Code (claude.ai/code) へのガイダンスを提供します。

> **ユーザー向け概要**: [README.md](./README.md) を参照してください

## プロジェクト概要

これは「共円」パズルゲーム用のGo製REST APIサーバーです。プレイヤーはグリッド上に石を配置し、ちょうど4つの石で円や直線を形成します。サーバーはCloud Run上で動作し、DatastoreモードのFirestore（既存Datastoreデータと互換）を永続化に使用します。

### コアゲームロジック
- **共円判定**: メインアルゴリズム（`models/kyouen.go`）が4つの石が有効な共円（円または直線）を形成するかを判定
- **ステージ検証**: 新しいステージは5個以上の石を持ち、少なくとも1つの有効な共円を含む必要があります
- **重複防止**: 回転と反転を含めたステージの重複チェック

## 開発コマンド

### ローカル開発
```bash
# Cloud Run対応サーバーの起動
go run cmd/server/main.go  # 本番Datastoreに接続
go run cmd/test_server/main.go  # Datastore接続テスト

# ローカルサーバーへのアクセス: http://localhost:8080/
```

### テスト
```bash
# 全テストの実行
go test -v ./...

# 特定パッケージのテスト実行
go test -v ./models
```

### ビルド
```bash
# アプリケーションのビルド
go build -v ./...
```

### 開発データ初期化
```bash
# Datastoreエミュレーターの起動
gcloud emulators firestore start --database-mode=datastore-mode --host-port=0.0.0.0:9098
# Firebase Authエミュレーターの起動
firebase emulators:start

# 開発環境に初期ステージデータをDatastoreエミュレーターに登録
DATASTORE_EMULATOR_HOST=localhost:9098 go run cmd/seed/main.go

# Datastoreエミュレーター使用時
DATASTORE_EMULATOR_HOST=localhost:9098 FIREBASE_AUTH_EMULATOR_HOST=localhost:9099 go run cmd/server/main.go
```

### OpenAPI コード生成
```bash
# OpenAPI仕様からGoモデルを生成（Docker経由）
docker run --rm -v $(pwd):/local openapitools/openapi-generator-cli generate -i /local/docs/specs/index.yaml -g go-server -o /local/tmp
cp tmp/go/* internal/generated/openapi/
rm internal/generated/openapi/api_*.go internal/generated/openapi/routers.go internal/generated/openapi/logger.go
rm -rf tmp

# Androidクライアントモデルを生成（Docker経由）
docker run --rm -v $(pwd):/local openapitools/openapi-generator-cli generate -i /local/docs/specs/index.yaml -g kotlin -o /local/tmp --additional-properties="packageName=hm.orz.chaos114.android.tumekyouen.network"
cp -r tmp/src/main/kotlin/hm/orz/chaos114/android/tumekyouen/network/models ../kyouen-android/app/src/main/java/hm/orz/chaos114/android/tumekyouen/network
rm -rf tmp
```

### デプロイ
```bash
# GitHub Actions経由でのデプロイ（推奨）
./scripts/deploy.sh dev           # DEV環境にサーバーをデプロイ（デフォルト）
./scripts/deploy.sh dev server    # DEV環境にサーバーをデプロイ
./scripts/deploy.sh dev seed      # DEV環境にSeedジョブをデプロイ
./scripts/deploy.sh prod          # 本番環境にサーバーをデプロイ（確認付き）

# 前提条件: GitHub CLI
brew install gh && gh auth login
```

### Swagger UI
```bash
# API ドキュメントのローカル表示
docker run -p 10000:8080 -v $(pwd)/docs:/usr/share/nginx/html/docs -e API_URL=http://localhost:10000/docs/specs/index.yaml swaggerapi/swagger-ui
```

## アーキテクチャ

### API構造
- **Version 2 API**: 全エンドポイントに `/v2` プレフィックス
- **ユーザー管理**: Twitter OAuth経由でのログイン（Userモデルにレガシートークン）
- **ステージ管理**: パズルステージのCRUD操作と検証
- **クリア追跡**: StageUserリレーションによるユーザー進捗追跡

### データモデル
- **KyouenPuzzle**: サイズ、ステージ文字列、作成者、登録日付を含むステージデータ
- **User**: クリアステージ数とレガシーOAuthトークンを持つユーザーアカウント
- **StageUser**: タイムスタンプ付きのユーザークリアを追跡する多対多リレーション
- **KyouenPuzzleSummary**: グローバル統計（カウント、最終更新日）

### ステージ表現
ステージは以下の文字列で表現されます：
- `"0"`: 空のセル
- `"1"`: 黒石（パズルの石）
- `"2"`: 白石（ユーザーの解答）

### データベース
- **DatastoreモードのFirestore**: 既存Datastoreデータと互換性を保持
- **プロジェクトID**: 
  - **DEV環境**: `api-project-732262258565`
  - **本番環境**: `my-android-server`
- **ローカル開発**: ファイルベースストレージのデータストアエミュレーターを使用

### 主要ハンドラー
- `handlers/v2/stages.go`: 共円検証付きのステージCRUD（Gin対応）
- `handlers/v2/statics.go`: グローバル統計（Gin対応）
- `services/datastore.go`: Datastoreサービス層
- `cmd/server/main.go`: Cloud Run用メインエントリーポイント

## 重要ファイル
- `models/kyouen.go`: 共円判定のコアゲームロジック
- `cmd/server/main.go`: Cloud Run本番用エントリーポイント
- `cmd/test_server/main.go`: Datastore接続テスト用サーバー
- `docs/specs/index.yaml`: OpenAPI仕様
- `Dockerfile`: Cloud Run用Dockerイメージ設定
- `scripts/deploy.sh`: Cloud Run手動デプロイスクリプト
- `.github/workflows/pr_validation.yml`: GitHub Actions CI設定
- `.github/workflows/deploy-dev.yml`: DEV環境自動デプロイ設定
- `.github/workflows/deploy-prod.yml`: 本番環境手動デプロイ設定
- `.github/workflows/deploy-common.yml`: 共通デプロイロジック
