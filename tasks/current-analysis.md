# 現状分析: 共円サーバー

## アーキテクチャ概要

### 現在の技術スタック
- **プラットフォーム**: Google App Engine
- **言語**: Go 1.x（古いバージョン）
- **ルーター**: Gorilla Mux
- **データベース**: Google Cloud Datastore
- **認証**: Twitter OAuth（anaconda ライブラリ使用）

### プロジェクト構造分析

#### メインアプリケーション (`main.go`)
```go
// 現在の構成
- appengine パッケージ使用（GAE専用）
- Gorilla Mux でルーティング
- /v2 プレフィックス
- CORS ミドルウェア（開発環境のみ）
- ポート 8080 でサーバー起動
```

#### データアクセス層 (`db/db.go`)
```go
// Datastore クライアント
- プロジェクトID: "my-android-server" (ハードコード)
- グローバル変数でクライアント管理
- シンプルな初期化パターン
```

#### 依存関係 (`go.mod`)
- **問題点**: Go modules 使用だが古い依存関係
- **App Engine専用**: `google.golang.org/appengine v1.4.0`
- **Datastore**: `cloud.google.com/go v0.36.0` (古いバージョン)
- **Twitter OAuth**: `github.com/ChimeraCoder/anaconda v2.0.0`
- **ルーター**: `github.com/gorilla/mux v1.7.0`

### API エンドポイント構造

#### 現在のルーティング
```
/v2/users/login          - ユーザーログイン (Twitter OAuth)
/v2/stages               - ステージ管理
/v2/stages/{stageNo}/clear - ステージクリア
/v2/statics              - 統計情報
```

## データモデル分析

### 予想されるエンティティ
1. **KyouenPuzzle**: パズルステージ
   - Size, StageString, CreatedBy, RegisterDate
2. **User**: ユーザー情報
   - ClearedStageNum, OAuth情報
3. **StageUser**: ユーザー×ステージのクリア状況
4. **KyouenPuzzleSummary**: 統計情報

## 移行の課題と対応

### 1. プラットフォーム依存の除去
**課題**: App Engine 固有の機能使用
```go
// 現在
"google.golang.org/appengine"
appengine.IsDevAppServer()

// 移行後
環境変数での開発・本番判定
```

### 2. ルーターの変更
**課題**: Gorilla Mux → Gin
```go
// 現在
r := mux.NewRouter().PathPrefix("/v2").Subrouter()
r.HandleFunc("/stages", handlers.StagesHandler)

// 移行後
r := gin.Default()
v2 := r.Group("/v2")
v2.GET("/stages", handlers.GetStages)
v2.POST("/stages", handlers.CreateStage)
```

### 3. データアクセス層の変更
**課題**: Datastore → Firestore
```go
// 現在
"cloud.google.com/go/datastore"
client *datastore.Client

// 移行後
"cloud.google.com/go/firestore"
client *firestore.Client
```

### 4. ハンドラー関数の変更
**課題**: http.Handler → gin.HandlerFunc
```go
// 現在
func StagesHandler(w http.ResponseWriter, r *http.Request)

// 移行後
func GetStages(client *firestore.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 実装
    }
}
```

## 移行優先度

### Phase 1: 高優先度
1. **データアクセス層**: Datastore → Firestore
2. **メインアプリケーション**: Gin フレームワーク導入
3. **ルーティング**: REST API 設計の見直し

### Phase 2: 中優先度
1. **ハンドラー**: 各エンドポイントの Gin 対応
2. **エラーハンドリング**: 統一的なエラー処理
3. **ミドルウェア**: CORS、ログ、認証

### Phase 3: 低優先度
1. **テスト**: Firestore エミュレータ対応
2. **Docker化**: Cloud Run 対応
3. **CI/CD**: デプロイパイプライン

## 互換性の維持

### API仕様
- `/v2` プレフィックスの維持
- レスポンス形式の互換性確保
- エラーレスポンスの統一

### 認証
- Twitter OAuth の継続サポート
- 既存トークンの移行対応

### データ
- 既存データの完全移行
- ID体系の維持（可能な限り）

## 次のステップ

1. **依存関係の更新**
   ```bash
   go get github.com/gin-gonic/gin@latest
   go get cloud.google.com/go/firestore@latest
   go get google.golang.org/api/iterator@latest
   ```

2. **ハンドラーの詳細分析**
   - 各ハンドラーの実装確認
   - 必要な変更点の洗い出し

3. **データモデルの Firestore 対応**
   - タグの追加
   - クエリ方法の確認

4. **テスト戦略の策定**
   - 既存テストの分析
   - 新しいテスト方針の決定