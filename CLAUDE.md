# CLAUDE.md

このファイルは、このリポジトリでコードを扱う際のClaude Code (claude.ai/code) へのガイダンスを提供します。

## プロジェクト概要

これは「共円」パズルゲーム用のGo製REST APIサーバーです。プレイヤーはグリッド上に石を配置し、ちょうど4つの石で円や直線を形成します。サーバーはCloud Run上で動作し、DatastoreモードのFirestore（既存Datastoreデータと互換）を永続化に使用します。

### コアゲームロジック
- **共円判定**: メインアルゴリズム（`models/kyouen.go`）が4つの石が有効な共円（円または直線）を形成するかを判定
- **ステージ検証**: 新しいステージは5個以上の石を持ち、少なくとも1つの有効な共円を含む必要があります
- **重複防止**: 回転と反転を含めたステージの重複チェック

## 開発コマンド

### ローカル開発
```bash
# データストアエミュレーターを使用したローカル開発サーバーの起動（レガシーApp Engine）
dev_appserver.py app.yaml --datastore_path=`pwd`/database/db.datastore -A my-android-server --support_datastore_emulator True --enable_host_checking=false

# 新しいCloud Run対応サーバーの起動
go run cmd/server/main.go  # 本番Datastoreに接続
go run cmd/demo_server/main.go  # デモデータで動作確認
go run cmd/test_server/main.go  # Datastore接続テスト

# ローカルサーバーへのアクセス: http://localhost:8080/
# 管理コンソールへのアクセス: http://localhost:8000/ (App Engineのみ)
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

### OpenAPI コード生成
```bash
# OpenAPI仕様からGoモデルを生成
openapi-generator generate -i docs/specs/index.yaml -g go-server -o ./tmp
cp tmp/go/model_*.go openapi
rm -rf tmp

# Androidクライアントモデルを生成
openapi-generator generate -i docs/specs/index.yaml -g kotlin -o ./tmp --additional-properties="packageName=hm.orz.chaos114.android.tumekyouen.network"
cp -r tmp/src/main/kotlin/hm/orz/chaos114/android/tumekyouen/network/models ../kyouen-android/app/src/main/java/hm/orz/chaos114/android/tumekyouen/network
rm -rf tmp
```

### デプロイ
```bash
# レガシーApp Engine環境へのデプロイ
gcloud app deploy --no-promote
gcloud app deploy dispatch.yaml

# 新しいCloud Run環境へのデプロイ
./scripts/deploy.sh  # 自動デプロイスクリプト

# 手動デプロイ
docker build -t gcr.io/my-android-server/kyouen-server:latest .
docker push gcr.io/my-android-server/kyouen-server:latest
gcloud run deploy kyouen-server --image gcr.io/my-android-server/kyouen-server:latest --region asia-northeast1

# Cloud Build使用（推奨）
gcloud builds submit --config cloudbuild.yaml
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
- **プロジェクトID**: `my-android-server`（services/datastore.goで設定）
- **ローカル開発**: ファイルベースストレージのデータストアエミュレーターを使用

### 主要ハンドラー
**レガシーハンドラー（App Engine）:**
- `handlers/stages_handler.go`: 共円検証付きのステージCRUD
- `handlers/stages/stage_no/clear_handler.go`: ステージ完了追跡
- `handlers/users/login_handler.go`: Twitter OAuth認証
- `handlers/statics_handler.go`: グローバル統計

**新しいハンドラー（Cloud Run + Gin）:**
- `handlers/v2/stages.go`: 共円検証付きのステージCRUD（Gin対応）
- `handlers/v2/statics.go`: グローバル統計（Gin対応）
- `services/datastore.go`: Datastoreサービス層
- `cmd/server/main.go`: Cloud Run用メインエントリーポイント

## 重要ファイル
- `models/kyouen.go`: 共円判定のコアゲームロジック
- `main.go`: レガシーApp Engineサーバー設定
- `cmd/server/main.go`: Cloud Run本番用エントリーポイント
- `cmd/demo_server/main.go`: 認証不要のデモサーバー
- `cmd/test_server/main.go`: Datastore接続テスト用サーバー
- `docs/specs/index.yaml`: OpenAPI仕様
- `app.yaml`: Google App Engine設定（レガシー）
- `Dockerfile`: Cloud Run用Dockerイメージ設定
- `cloudbuild.yaml`: Cloud Build自動デプロイ設定
- `scripts/deploy.sh`: Cloud Run手動デプロイスクリプト
- `tasks/datastore-mode-migration.md`: 移行戦略ドキュメント
- `tasks/migration-plan.md`: 完了した移行計画ドキュメント