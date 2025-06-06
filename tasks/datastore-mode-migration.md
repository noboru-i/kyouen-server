# Datastoreモード対応移行計画

## 背景

既存データがDatastoreモードのFirestoreに存在するため、移行戦略を以下に変更します：

### Datastoreモード vs ネイティブモード

| 項目 | Datastoreモード | ネイティブモード |
|------|----------------|------------------|
| APIクライアント | `cloud.google.com/go/datastore` | `cloud.google.com/go/firestore` |
| データ構造 | Entity/Key ベース | Document/Collection ベース |
| クエリ構文 | Datastore Query | Firestore Query |
| 既存データ | ✅ 利用可能 | ❌ 移行が必要 |

## 新しい移行戦略

### Phase 1: ハイブリッド構成 ✅（部分完了）
- ✅ Gin フレームワークの導入
- ✅ 新しいアーキテクチャの基盤構築
- 🔄 **既存Datastoreクライアントの継続利用**

### Phase 2: サービス層統合
- 既存の `db/db.go` を活用
- Datastoreクライアントを Gin ハンドラーに統合
- 既存のエンティティ構造を維持

### Phase 3: 段階的なAPI移行
1. StaticsHandler: 既存 `db.KyouenPuzzleSummary` 利用
2. StagesHandler: 既存 `db.KyouenPuzzle` 利用
3. UsersHandler: 既存 `db.User` 利用
4. ClearHandler: 既存 `db.StageUser` 利用

### Phase 4: Cloud Run対応
- Docker化
- 環境変数での設定管理
- ヘルスチェック

## 実装アプローチ

### 1. 既存Datastoreサービスの統合

```go
// services/datastore.go
type DatastoreService struct {
    client *datastore.Client
}

func NewDatastoreService(projectID string) (*DatastoreService, error) {
    // 既存の db.InitDB() ロジックを活用
}
```

### 2. Ginハンドラーでの既存エンティティ利用

```go
// handlers/v2/statics.go
func GetStatics(datastoreService *DatastoreService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 既存の db.KyouenPuzzleSummary を利用
        summary, err := datastoreService.GetSummary()
        // ...
    }
}
```

### 3. 既存ビジネスロジックの保持

```go
// 既存の共円判定ロジックはそのまま利用
import "kyouen-server/models"

stage := models.NewKyouenStage(size, stageString)
if !stage.HasKyouen() {
    // バリデーションエラー
}
```

## メリット

1. **データ移行不要**: 既存データをそのまま利用
2. **段階的移行**: リスクを最小化
3. **ビジネスロジック保持**: 共円判定などの核心機能は変更なし
4. **パフォーマンス**: データ移行のオーバーヘッドなし

## 技術的変更点

### 依存関係
- `cloud.google.com/go/datastore` を継続利用
- 新規の `cloud.google.com/go/firestore` は不要
- Gin関連の依存関係は維持

### アーキテクチャ
```
main.go (Gin) 
  ↓
handlers/v2/ (Gin handlers)
  ↓
services/datastore.go (新規)
  ↓
db/db.go (既存Datastoreクライアント)
  ↓
Google Cloud Datastore (既存データ)
```

### ファイル構成
```
kyouen-server/
├── main_v2.go              # 新しいGinアプリケーション
├── db/                     # 既存Datastoreクライアント（継続利用）
│   ├── db.go
│   └── models.go
├── services/
│   └── datastore.go        # 新規：Datastoreサービス層
├── handlers/v2/            # 新規：Ginハンドラー
│   ├── statics.go
│   ├── stages.go
│   └── users.go
├── models/                 # 既存ビジネスロジック（継続利用）
│   ├── kyouen.go
│   ├── line.go
│   └── point.go
└── config/                 # 新規：設定管理
    └── config.go
```

## 次のアクション

1. **services/datastore.go** の作成
2. **handlers/v2/statics.go** の実装（既存データ利用）
3. **既存データでの動作確認**
4. **段階的なハンドラー移行**

この方針により、既存データを活用しながら新しいアーキテクチャに移行できます。