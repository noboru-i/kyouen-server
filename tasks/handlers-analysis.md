# ハンドラー詳細分析

## 各ハンドラーの現状と移行計画

### 1. StagesHandler (`handlers/stages_handler.go`)

#### 現在の実装
- **GET /v2/stages**: ステージ一覧取得
  - クエリパラメータ: `start_stage_no`, `limit` (最大100)
  - Datastore クエリ: `Filter("stageNo >=", param.startStageNo).Limit(param.limit)`
  
- **POST /v2/stages**: 新しいステージ作成
  - バリデーション:
    - 石の数が5個以上
    - 共円判定の実行
    - 重複チェック（回転・反転含む）
  - 自動採番: 最大stageNo + 1
  - 統計情報の更新

#### Gin移行時の変更点
```go
// 現在
func StagesHandler(w http.ResponseWriter, r *http.Request)

// 移行後
func GetStages(client *firestore.Client) gin.HandlerFunc
func CreateStage(client *firestore.Client) gin.HandlerFunc
```

#### Firestore移行時の変更点
- クエリ構文の変更
- エンティティ取得方法の変更
- トランザクション処理の調整

### 2. LoginHandler (`handlers/users/login_handler.go`)

#### 現在の実装
- **POST /v2/users/login**: Twitter OAuth ログイン
- 依存関係:
  - `github.com/ChimeraCoder/anaconda`: Twitter API
  - `firebase.google.com/go`: Firebase Auth（カスタムトークン生成）
- 処理フロー:
  1. Twitter API でユーザー認証
  2. ユーザー情報の upsert
  3. Firebase カスタムトークン生成
  4. レスポンス返却

#### 移行時の課題
- **Firebase サービスアカウントキー**: ハードコードされたファイルパス
- **環境変数**: `CONSUMER_KEY`, `CONSUMER_SECRET`
- **Datastore キー生成**: `"KEY"+strconv.FormatInt(user.Id, 10)`

#### 移行後の改善点
```go
// 現在（問題のあるコード）
opt := option.WithCredentialsFile("api-project-1046368181881-firebase-adminsdk-df1u6-039d87ad7a.json")

// 移行後
opt := option.WithCredentialsFile(os.Getenv("FIREBASE_CREDENTIALS_FILE"))
```

### 3. ClearHandler (`handlers/stages/stage_no/clear_handler.go`)

#### 現在の実装
- **POST /v2/stages/{stageNo}/clear**: ステージクリア記録
- 認証: Firebase ID トークン検証
- 処理フロー:
  1. ステージ番号の取得
  2. クリアデータの検証（共円判定）
  3. 既存ステージとの一致確認
  4. ユーザー認証
  5. StageUser エンティティの作成/更新
  6. ユーザーのクリア数更新

#### 移行時の課題
- **Gorilla Mux**: パスパラメータ取得方法
- **Firebase認証**: サービスアカウントキーのハードコード
- **複雑なクエリ**: 複数フィルター条件

#### Gin移行時の変更
```go
// 現在
stageNo, err := strconv.Atoi(mux.Vars(r)["stageNo"])

// 移行後
stageNo, err := strconv.Atoi(c.Param("stageNo"))
```

### 4. StaticsHandler (`handlers/statics_handler.go`)

#### 現在の実装
- **GET /v2/statics**: 統計情報取得
- シンプルな実装: KyouenPuzzleSummary エンティティの取得

#### 移行時の変更点
- 最小限の変更で済む
- Firestore クエリへの変更のみ

## 共通の移行課題

### 1. エラーハンドリング
**現在**: パニックまたは直接レスポンス書き込み
```go
if err != nil {
    panic("database error." + err.Error())
}
```

**移行後**: 統一されたエラーレスポンス
```go
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
}
```

### 2. 認証ミドルウェア
Firebase認証の共通化が必要
```go
func FirebaseAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Firebase ID トークン検証
        // ユーザー情報をコンテキストに設定
        c.Next()
    }
}
```

### 3. CORS設定
現在は開発環境のみ対応
```go
// 移行後: Gin CORS ミドルウェア使用
r.Use(cors.Default())
```

### 4. ログ出力
統一されたログ形式が必要
```go
// 構造化ログの導入
log.Printf("User login: userID=%s, screenName=%s", user.ID, user.ScreenName)
```

## 移行順序

### Phase 1: 基盤整備
1. Gin フレームワークの導入
2. Firestore クライアントの設定
3. 共通ミドルウェアの実装

### Phase 2: シンプルなハンドラーから移行
1. **StaticsHandler**: 最もシンプル
2. **StagesHandler (GET)**: 読み取り専用
3. **StagesHandler (POST)**: 複雑なビジネスロジック

### Phase 3: 認証が関わるハンドラー
1. **LoginHandler**: 認証基盤の整備
2. **ClearHandler**: 認証 + 複雑なロジック

## テスト戦略

### 単体テスト
- Firestore エミュレータの使用
- モックを使った認証テスト
- ビジネスロジックの分離テスト

### 統合テスト
- エンドツーエンドのAPIテスト
- 認証フローのテスト
- データ整合性のテスト

## 設定ファイルの整理

### 環境変数
```bash
# Twitter OAuth
CONSUMER_KEY=xxx
CONSUMER_SECRET=xxx

# Firebase
FIREBASE_CREDENTIALS_FILE=path/to/service-account.json
GOOGLE_CLOUD_PROJECT=my-android-server

# Server
PORT=8080
GIN_MODE=release  # or debug
```

### 設定構造体
```go
type Config struct {
    Port        string
    ProjectID   string
    TwitterConfig TwitterConfig
    FirebaseConfig FirebaseConfig
}
```