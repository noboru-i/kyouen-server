# 共円サーバー ドキュメント

このディレクトリには、共円パズルゲームAPIサーバーの技術ドキュメントが含まれています。

## ファイル構成

### データベーススキーマ
- **[datastore-schema.json](./datastore-schema.json)** - Datastoreエンティティのスキーマ定義（JSON Schema形式）

### API仕様
- **[specs/index.yaml](./specs/index.yaml)** - OpenAPI 3.0仕様（REST API定義）

## Datastoreスキーマドキュメント

`datastore-schema.json`は、プロジェクトで使用されているFirestore（Datastoreモード）のエンティティ構造を定義しています。

### 含まれる情報

#### エンティティ定義
- **KyouenPuzzleSummary** - グローバル統計情報
- **KyouenPuzzle** - パズルステージデータ
- **User** - ユーザーアカウント情報  
- **StageUser** - ユーザーとステージの多対多関係

#### 各エンティティの詳細
- フィールド定義（型、制約、説明）
- キーパターン（自動生成 vs 名前付きキー）
- インデックス設定
- エンティティ間のリレーション
- 利用パターンとクエリ例

#### 追加情報
- ゲームロジックの説明
- バリデーションルール
- 環境設定（DEV/本番）
- 重複検出ロジック

### 使用方法

#### 開発者向け
- 新機能開発時のデータモデル参照
- クエリ設計時のインデックス確認
- エンティティ関係の理解

#### インフラ担当者向け
- Datastore設定の確認
- インデックス管理
- 容量計画とコスト見積もり

#### JSON Schemaツールとの連携
この定義は標準的なJSON Schema（draft-07）形式なので、以下のツールで活用できます：

```bash
# スキーマの検証
npm install -g ajv-cli
ajv validate -s datastore-schema.json -d sample-data.json

# ドキュメント生成
npm install -g json-schema-to-markdown
json-schema-to-markdown -s datastore-schema.json -o schema-docs.md
```

## 関連ドキュメント

- [CLAUDE.md](../CLAUDE.md) - 開発ガイダンス
- [README.md](../README.md) - プロジェクト概要
- [OpenAPI仕様](./specs/index.yaml) - REST API詳細

## 更新について

データモデルに変更があった場合は、以下を更新してください：

1. `internal/datastore/models.go` - Goの構造体定義
2. `docs/datastore-schema.json` - このスキーマ定義
3. 必要に応じてマイグレーション手順をドキュメント化

スキーマ定義の一貫性を保つため、コードレビュー時は両方のファイルをチェックしてください。