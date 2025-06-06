# CLAUDE.md

このファイルは、このリポジトリでコードを扱う際のClaude Code (claude.ai/code) へのガイダンスを提供します。

## プロジェクト概要

これは「共円」パズルゲーム用のGo製REST APIサーバーです。プレイヤーはグリッド上に石を配置し、ちょうど4つの石で円や直線を形成します。サーバーはGoogle App Engine上で動作し、Google Cloud Datastoreを永続化に使用します。

### コアゲームロジック
- **共円判定**: メインアルゴリズム（`models/kyouen.go`）が4つの石が有効な共円（円または直線）を形成するかを判定
- **ステージ検証**: 新しいステージは5個以上の石を持ち、少なくとも1つの有効な共円を含む必要があります
- **重複防止**: 回転と反転を含めたステージの重複チェック

## 開発コマンド

### ローカル開発
```bash
# データストアエミュレーターを使用したローカル開発サーバーの起動
dev_appserver.py app.yaml --datastore_path=`pwd`/database/db.datastore -A my-android-server --support_datastore_emulator True --enable_host_checking=false

# ローカルサーバーへのアクセス: http://localhost:8080/
# 管理コンソールへのアクセス: http://localhost:8000/
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
# 本番環境へのデプロイ（テスト用にno-promote）
gcloud app deploy --no-promote

# dispatch設定のデプロイ
gcloud app deploy dispatch.yaml
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
- **Google Cloud Datastore**: メインの永続化レイヤー
- **プロジェクトID**: `my-android-server`（db/db.goにハードコード）
- **ローカル開発**: ファイルベースストレージのデータストアエミュレーターを使用

### 主要ハンドラー
- `handlers/stages_handler.go`: 共円検証付きのステージCRUD
- `handlers/stages/stage_no/clear_handler.go`: ステージ完了追跡
- `handlers/users/login_handler.go`: Twitter OAuth認証
- `handlers/statics_handler.go`: グローバル統計

## 重要ファイル
- `models/kyouen.go`: 共円判定のコアゲームロジック
- `main.go`: ルーティングとCORSミドルウェアを含むサーバー設定
- `docs/specs/index.yaml`: OpenAPI仕様
- `app.yaml`: Google App Engine設定
- `secret.yaml`: OAuthキー用の環境変数（設定が必要）