# 実装計画書: Cloud Run + Firestore 移行

## 概要
共円サーバーをApp Engine + DatastoreからCloud Run + Firestoreに移行するための段階的実装計画

## フェーズ別実装計画

### Phase 1: 基盤構築 📋
**目標**: 新しい技術スタックの基盤整備
**期間**: 1-2日

#### 1.1 依存関係の更新
```bash
# 新しい依存関係の追加
go get github.com/gin-gonic/gin@latest
go get cloud.google.com/go/firestore@latest
go get google.golang.org/api/iterator@latest
go get github.com/stretchr/testify/assert@latest

# レガシー依存関係の整理
# App Engine関連の依存関係を段階的に削除
```

#### 1.2 設定管理の整備
```go
// config/config.go
type Config struct {
    Port           string
    ProjectID      string
    Environment    string
    TwitterConfig  TwitterConfig
    FirebaseConfig FirebaseConfig
}
```

#### 1.3 新しいmain.goの作成
```go
// main_new.go (並行開発用)
func main() {
    // Firestore クライアント初期化
    // Gin エンジン設定
    // ルーティング設定
    // サーバー起動
}
```

### Phase 2: データアクセス層移行 🗄️
**目標**: Datastore → Firestore への完全移行
**期間**: 2-3日

#### 2.1 Firestore サービス層作成
```go
// services/firestore.go
type FirestoreService struct {
    client *firestore.Client
}

func (s *FirestoreService) GetStages(limit int, startAfter string) ([]models.Stage, error)
func (s *FirestoreService) CreateStage(stage *models.Stage) (*models.Stage, error)
func (s *FirestoreService) GetUser(userID string) (*models.User, error)
// ... その他のCRUD操作
```

#### 2.2 モデルのFirestore対応
```go
// models/firestore/stage.go
type Stage struct {
    ID         string    `firestore:"-" json:"id"`
    StageNo    int64     `firestore:"stageNo" json:"stageNo"`
    Size       int64     `firestore:"size" json:"size"`
    Stage      string    `firestore:"stage" json:"stage"`
    Creator    string    `firestore:"creator" json:"creator"`
    RegistDate time.Time `firestore:"registDate" json:"registDate"`
    CreatedAt  time.Time `firestore:"createdAt" json:"createdAt"`
    UpdatedAt  time.Time `firestore:"updatedAt" json:"updatedAt"`
}
```

#### 2.3 データマイグレーション準備
```go
// tools/migrate/datastore_to_firestore.go
// Datastoreからデータを読み取り、Firestoreに移行するツール
```

### Phase 3: API層の移行 🌐
**目標**: ハンドラーのGin対応とFirestore連携
**期間**: 3-4日

#### 3.1 シンプルなエンドポイントから移行

##### 3.1.1 StaticsHandler
```go
// handlers/v2/statics.go
func GetStatics(firestoreService *services.FirestoreService) gin.HandlerFunc {
    return func(c *gin.Context) {
        stats, err := firestoreService.GetStatics()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, stats)
    }
}
```

##### 3.1.2 StagesHandler (GET)
```go
// handlers/v2/stages.go
func GetStages(firestoreService *services.FirestoreService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // クエリパラメータの解析
        startStageNo := c.DefaultQuery("start_stage_no", "0")
        limit := c.DefaultQuery("limit", "10")
        
        // ビジネスロジック
        stages, err := firestoreService.GetStages(limit, startStageNo)
        // レスポンス
        c.JSON(http.StatusOK, stages)
    }
}
```

#### 3.2 複雑なエンドポイントの移行

##### 3.2.1 StagesHandler (POST)
```go
func CreateStage(firestoreService *services.FirestoreService) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req models.CreateStageRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        
        // 既存の共円判定ロジックを活用
        stage := models.NewKyouenStage(int(req.Size), req.Stage)
        if !validateStage(stage) {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stage"})
            return
        }
        
        // 保存処理
        result, err := firestoreService.CreateStage(&req)
        c.JSON(http.StatusCreated, result)
    }
}
```

### Phase 4: 認証システムの移行 🔐
**目標**: Firebase Auth統合とセキュリティ強化
**期間**: 2-3日

#### 4.1 認証ミドルウェアの作成
```go
// middleware/auth.go
func FirebaseAuthMiddleware(firebaseService *services.FirebaseService) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractBearerToken(c)
        user, err := firebaseService.VerifyToken(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        c.Set("user", user)
        c.Next()
    }
}
```

#### 4.2 LoginHandler の移行
```go
// handlers/v2/auth.go
func Login(services *services.Services) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req models.LoginRequest
        // Twitter OAuth検証
        // Firebase カスタムトークン生成
        // ユーザー情報のupsert
        c.JSON(http.StatusOK, response)
    }
}
```

#### 4.3 ClearHandler の移行
```go
func ClearStage(services *services.Services) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := c.MustGet("user").(*models.User)
        stageNo := c.Param("stageNo")
        
        var req models.ClearStageRequest
        // 共円判定
        // ステージ検証
        // クリア記録の保存
        c.JSON(http.StatusOK, gin.H{"status": "success"})
    }
}
```

### Phase 5: Docker化とCloud Run対応 🐳
**目標**: コンテナ化とクラウド対応
**期間**: 1-2日

#### 5.1 Dockerfile作成
```dockerfile
# Multi-stage build
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

#### 5.2 Cloud Run設定
```yaml
# cloudbuild.yaml
steps:
  - name: 'gcr.io/cloud-builders/docker'
    args: ['build', '-t', 'gcr.io/$PROJECT_ID/kyouen-server:$COMMIT_SHA', '.']
  - name: 'gcr.io/cloud-builders/docker'
    args: ['push', 'gcr.io/$PROJECT_ID/kyouen-server:$COMMIT_SHA']
  - name: 'gcr.io/google.com/cloudsdktool/cloud-sdk'
    entrypoint: 'gcloud'
    args: ['run', 'deploy', 'kyouen-server', '--image', 'gcr.io/$PROJECT_ID/kyouen-server:$COMMIT_SHA', '--region', 'asia-northeast1']
```

### Phase 6: テスト実装 🧪
**目標**: 包括的なテストスイートの構築
**期間**: 2-3日

#### 6.1 テストユーティリティ
```go
// testutils/firestore.go
func SetupFirestoreTest() *firestore.Client {
    os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080")
    client, _ := firestore.NewClient(context.Background(), "test-project")
    return client
}
```

#### 6.2 ハンドラーテスト
```go
// handlers/v2/stages_test.go
func TestGetStages(t *testing.T) {
    // Firestoreエミュレータを使用したテスト
    // Ginのテストモードでのリクエスト/レスポンステスト
}
```

### Phase 7: 本番移行 🚀
**目標**: 安全な本番環境移行
**期間**: 1-2日

#### 7.1 データ移行
- Datastoreからの完全データエクスポート
- Firestoreへのデータインポート
- データ整合性の検証

#### 7.2 DNS切り替え
- 新しいCloud Runサービスの本番デプロイ
- ロードバランサーの設定
- 段階的なトラフィック移行

## リスク管理と対策

### 高リスク項目
1. **データ移行時の整合性**: バックアップとロールバック計画
2. **API互換性**: 詳細なE2Eテスト
3. **認証フローの中断**: 段階的移行とフォールバック

### 対策
- 本番環境での並行運用期間を設ける
- 詳細なモニタリングとアラート設定
- 即座にロールバック可能な体制

## 開発環境構築

### ローカル開発
```bash
# Firestoreエミュレータ起動
firebase emulators:start --only firestore

# アプリケーション起動
go run main.go

# テスト実行
go test ./... -v
```

### 環境変数
```bash
# .env.local
PORT=8080
GOOGLE_CLOUD_PROJECT=my-android-server
FIRESTORE_EMULATOR_HOST=localhost:8080
CONSUMER_KEY=xxx
CONSUMER_SECRET=xxx
FIREBASE_CREDENTIALS_FILE=path/to/service-account.json
```

## 成果物

### 実装ファイル
1. `main.go` - 新しいメインアプリケーション
2. `services/` - Firestoreサービス層
3. `handlers/v2/` - 新しいGinハンドラー
4. `middleware/` - 認証・CORS等のミドルウェア
5. `models/firestore/` - Firestore対応モデル
6. `config/` - 設定管理
7. `Dockerfile` - コンテナ設定

### 設定ファイル
1. `firebase.json` - Firestore設定
2. `cloudbuild.yaml` - CI/CD設定
3. `.dockerignore` - Docker無視ファイル

### ドキュメント
1. API仕様書の更新
2. デプロイメント手順書
3. 運用マニュアル

## 次のアクション
1. Phase 1の依存関係更新から開始
2. 各フェーズごとの進捗確認
3. 問題発生時の対応手順の確認