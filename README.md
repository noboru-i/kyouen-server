# 共円パズルゲーム API サーバー

「共円」パズルゲーム用のREST APIサーバーです。プレイヤーはグリッド上に石を配置し、ちょうど4つの石で円や直線を形成する知的パズルゲームをお楽しみいただけます。

> **開発者向け詳細ガイド**: [CLAUDE.md](./CLAUDE.md) を参照してください

## 🏗️ アーキテクチャ

- **プラットフォーム**: Cloud Run (コンテナベース)
- **フレームワーク**: Gin (Go製高速Webフレームワーク)
- **データベース**: DatastoreモードFirestore (既存データと互換性を保持)
- **言語**: Go 1.23+
- **認証**: Twitter OAuth + Firebase Authentication

## 🚀 クイックスタート

### 前提条件

- Go 1.23以上
- GitHub CLI (`gh`) - デプロイスクリプト用

### ローカル開発

#### デモサーバー（認証不要）
```bash
go run cmd/demo_server/main.go
```
サンプルデータでAPI動作を確認できます。

#### 本番接続サーバー
```bash
go run cmd/server/main.go
```
実際のDatastoreに接続して動作します。

**アクセス先:** http://localhost:8080/

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

## 🚢 デプロイメント

### Cloud Runへのデプロイ

```bash
# DEV環境にデプロイ（デフォルト）
./scripts/deploy.sh dev

# 本番環境にデプロイ（確認付き）
./scripts/deploy.sh prod
```

> **詳細な開発・デプロイ手順**: [CLAUDE.md](./CLAUDE.md) を参照してください


## 🧪 テスト

```bash
# 全テスト実行
go test -v ./...

# ビルドテスト
go build -v ./...
```

## 🚀 CI/CD

GitHub Actionsによる自動CI/CDを設定済み：
- **PR検証**: Go 1.23での自動テスト・ビルド
- **自動デプロイ**: DEV環境（mainブランチ）、本番環境（手動実行）


## ⚙️ 環境設定

```bash
# .env.example を .env にコピーしてローカル環境用設定
cp .env.example .env
```

## 🤝 開発について

詳細な開発ガイドは [CLAUDE.md](./CLAUDE.md) を参照してください。