# Go + Cloud Run + Firestore 開発ガイド

## 概要

このガイドでは、Go言語を使用してCloud Run上でHTTP APIを構築し、Firestoreをデータベースとして使用する方法について説明します。自動テストを重視した開発手法とClaude Codeでの効率的な開発について解説します。

## プロジェクト構成

```
my-api-project/
├── main.go
├── go.mod
├── go.sum
├── Dockerfile
├── .dockerignore
├── .gitignore
├── handlers/
│   ├── user.go
│   ├── user_test.go
│   ├── post.go
│   └── post_test.go
├── models/
│   ├── user.go
│   └── post.go
├── services/
│   ├── firestore.go
│   └── firestore_test.go
├── testutils/
│   └── setup.go
├── firebase.json
├── .firebaserc
└── cloudbuild.yaml
```

## 1. 初期セットアップ

### Go モジュール初期化

```bash
go mod init my-api-project
go get github.com/gin-gonic/gin
go get cloud.google.com/go/firestore
go get github.com/stretchr/testify
go get google.golang.org/api/iterator
```

### Firebase設定

```bash
npm install -g firebase-tools
firebase init firestore
```

## 2. メインアプリケーション

### main.go

```go
package main

import (
    "context"
    "log"
    "os"

    "cloud.google.com/go/firestore"
    "github.com/gin-gonic/gin"
    "my-api-project/handlers"
    "my-api-project/services"
)

type App struct {
    FirestoreClient *firestore.Client
}

func main() {
    ctx := context.Background()
    
    // Firestore クライアント初期化
    projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
    if projectID == "" {
        projectID = "your-project-id" // デフォルト値
    }
    
    client, err := firestore.NewClient(ctx, projectID)
    if err != nil {
        log.Fatalf("Failed to create Firestore client: %v", err)
    }
    defer client.Close()
    
    // アプリケーション初期化
    app := &App{
        FirestoreClient: client,
    }
    
    // Ginエンジン設定
    r := app.setupRoutes()
    
    // サーバー起動
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("Server starting on port %s", port)
    r.Run(":" + port)
}

func (app *App) setupRoutes() *gin.Engine {
    r := gin.Default()
    
    // ヘルスチェック
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    // API v1 グループ
    v1 := r.Group("/api/v1")
    {
        // ユーザー関連API
        users := v1.Group("/users")
        {
            users.GET("", handlers.GetUsers(app.FirestoreClient))
            users.POST("", handlers.CreateUser(app.FirestoreClient))
            users.GET("/:id", handlers.GetUser(app.FirestoreClient))
            users.PUT("/:id", handlers.UpdateUser(app.FirestoreClient))
            users.DELETE("/:id", handlers.DeleteUser(app.FirestoreClient))
        }
        
        // 投稿関連API
        posts := v1.Group("/posts")
        {
            posts.GET("", handlers.GetPosts(app.FirestoreClient))
            posts.POST("", handlers.CreatePost(app.FirestoreClient))
            posts.GET("/:id", handlers.GetPost(app.FirestoreClient))
            posts.PUT("/:id", handlers.UpdatePost(app.FirestoreClient))
            posts.DELETE("/:id", handlers.DeletePost(app.FirestoreClient))
        }
    }
    
    return r
}
```

## 3. モデル定義

### models/user.go

```go
package models

import "time"

type User struct {
    ID        string    `firestore:"-" json:"id"`
    Name      string    `firestore:"name" json:"name" binding:"required"`
    Email     string    `firestore:"email" json:"email" binding:"required,email"`
    CreatedAt time.Time `firestore:"created_at" json:"created_at"`
    UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}

type CreateUserRequest struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

type UpdateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email" binding:"omitempty,email"`
}
```

### models/post.go

```go
package models

import "time"

type Post struct {
    ID        string    `firestore:"-" json:"id"`
    Title     string    `firestore:"title" json:"title" binding:"required"`
    Content   string    `firestore:"content" json:"content" binding:"required"`
    UserID    string    `firestore:"user_id" json:"user_id" binding:"required"`
    CreatedAt time.Time `firestore:"created_at" json:"created_at"`
    UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}

type CreatePostRequest struct {
    Title   string `json:"title" binding:"required"`
    Content string `json:"content" binding:"required"`
    UserID  string `json:"user_id" binding:"required"`
}

type UpdatePostRequest struct {
    Title   string `json:"title"`
    Content string `json:"content"`
}
```

## 4. ハンドラー実装

### handlers/user.go

```go
package handlers

import (
    "context"
    "net/http"
    "time"

    "cloud.google.com/go/firestore"
    "github.com/gin-gonic/gin"
    "google.golang.org/api/iterator"
    "my-api-project/models"
)

func GetUsers(client *firestore.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := context.Background()
        
        iter := client.Collection("users").Documents(ctx)
        var users []models.User
        
        for {
            doc, err := iter.Next()
            if err == iterator.Done {
                break
            }
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
            
            var user models.User
            if err := doc.DataTo(&user); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
            user.ID = doc.Ref.ID
            users = append(users, user)
        }
        
        c.JSON(http.StatusOK, users)
    }
}

func CreateUser(client *firestore.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req models.CreateUserRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        
        ctx := context.Background()
        now := time.Now()
        
        user := models.User{
            Name:      req.Name,
            Email:     req.Email,
            CreatedAt: now,
            UpdatedAt: now,
        }
        
        docRef, _, err := client.Collection("users").Add(ctx, user)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        
        user.ID = docRef.ID
        c.JSON(http.StatusCreated, user)
    }
}

func GetUser(client *firestore.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        ctx := context.Background()
        
        doc, err := client.Collection("users").Doc(id).Get(ctx)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
            return
        }
        
        var user models.User
        if err := doc.DataTo(&user); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        user.ID = doc.Ref.ID
        
        c.JSON(http.StatusOK, user)
    }
}

func UpdateUser(client *firestore.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        var req models.UpdateUserRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        
        ctx := context.Background()
        
        // 更新データ準備
        updates := []firestore.Update{
            {Path: "updated_at", Value: time.Now()},
        }
        
        if req.Name != "" {
            updates = append(updates, firestore.Update{Path: "name", Value: req.Name})
        }
        if req.Email != "" {
            updates = append(updates, firestore.Update{Path: "email", Value: req.Email})
        }
        
        _, err := client.Collection("users").Doc(id).Update(ctx, updates)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        
        c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
    }
}

func DeleteUser(client *firestore.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        ctx := context.Background()
        
        _, err := client.Collection("users").Doc(id).Delete(ctx)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        
        c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
    }
}
```

## 5. テスト設定

### testutils/setup.go

```go
package testutils

import (
    "context"
    "log"
    "os"

    "cloud.google.com/go/firestore"
    "google.golang.org/api/iterator"
)

func SetupFirestoreTest() *firestore.Client {
    // エミュレータホスト設定
    os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080")
    
    ctx := context.Background()
    client, err := firestore.NewClient(ctx, "test-project")
    if err != nil {
        log.Fatal(err)
    }
    return client
}

func CleanupFirestore(client *firestore.Client) {
    ctx := context.Background()
    collections := []string{"users", "posts"}
    
    for _, collection := range collections {
        iter := client.Collection(collection).Documents(ctx)
        batch := client.Batch()
        
        for {
            doc, err := iter.Next()
            if err == iterator.Done {
                break
            }
            if err != nil {
                log.Printf("Error getting document: %v", err)
                continue
            }
            batch.Delete(doc.Ref)
        }
        
        if _, err := batch.Commit(ctx); err != nil {
            log.Printf("Error cleaning up collection %s: %v", collection, err)
        }
    }
}
```

### handlers/user_test.go

```go
package handlers

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "my-api-project/models"
    "my-api-project/testutils"
)

func TestCreateUser(t *testing.T) {
    // Firestoreエミュレータセットアップ
    client := testutils.SetupFirestoreTest()
    defer client.Close()
    defer testutils.CleanupFirestore(client)
    
    // Ginエンジンセットアップ
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.POST("/users", CreateUser(client))
    
    // テストデータ
    user := models.CreateUserRequest{
        Name:  "テストユーザー",
        Email: "test@example.com",
    }
    
    jsonData, _ := json.Marshal(user)
    req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var response models.User
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, user.Name, response.Name)
    assert.Equal(t, user.Email, response.Email)
    assert.NotEmpty(t, response.ID)
}

func TestGetUsers(t *testing.T) {
    client := testutils.SetupFirestoreTest()
    defer client.Close()
    defer testutils.CleanupFirestore(client)
    
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.GET("/users", GetUsers(client))
    
    req, _ := http.NewRequest("GET", "/users", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    
    var users []models.User
    err := json.Unmarshal(w.Body.Bytes(), &users)
    assert.NoError(t, err)
    assert.IsType(t, []models.User{}, users)
}
```

## 6. Docker設定

### Dockerfile

```dockerfile
# ビルドステージ
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 実行ステージ
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
```

### .dockerignore

```
.git
.gitignore
README.md
Dockerfile
.dockerignore
node_modules
firebase-debug.log
.firebase/
```

## 7. Firebase設定

### firebase.json

```json
{
  "emulators": {
    "firestore": {
      "port": 8080
    },
    "ui": {
      "enabled": true,
      "port": 4000
    }
  }
}
```

## 8. Cloud Build設定

### cloudbuild.yaml

```yaml
steps:
  # Go テスト実行 (Firestoreエミュレータ使用)
  - name: 'gcr.io/google.com/cloudsdktool/cloud-sdk'
    entrypoint: 'bash'
    args:
      - '-c'
      - |
        gcloud emulators firestore start --host-port=0.0.0.0:8080 &
        sleep 10
        export FIRESTORE_EMULATOR_HOST=localhost:8080
        go test ./... -v
    env:
      - 'FIRESTORE_EMULATOR_HOST=localhost:8080'

  # Docker イメージビルド
  - name: 'gcr.io/cloud-builders/docker'
    args: ['build', '-t', 'gcr.io/$PROJECT_ID/my-api:$COMMIT_SHA', '.']

  # Docker イメージプッシュ
  - name: 'gcr.io/cloud-builders/docker'
    args: ['push', 'gcr.io/$PROJECT_ID/my-api:$COMMIT_SHA']

  # Cloud Run デプロイ
  - name: 'gcr.io/google.com/cloudsdktool/cloud-sdk'
    entrypoint: 'gcloud'
    args:
      - 'run'
      - 'deploy'
      - 'my-api'
      - '--image'
      - 'gcr.io/$PROJECT_ID/my-api:$COMMIT_SHA'
      - '--region'
      - 'asia-northeast1'
      - '--platform'
      - 'managed'
      - '--allow-unauthenticated'

images:
  - 'gcr.io/$PROJECT_ID/my-api:$COMMIT_SHA'
```

## 9. 開発ワークフロー

### ローカル開発

```bash
# Firestoreエミュレータ起動
firebase emulators:start --only firestore

# 別ターミナルでアプリケーション起動
go run main.go

# テスト実行
go test ./... -v
```

### Claude Codeでの開発のコツ

1. **段階的な実装**: 1つずつAPIエンドポイントを実装してテスト
2. **テストファースト**: ハンドラーの実装前にテストケースを書く
3. **エラーハンドリング**: 各ステップでエラーケースを考慮
4. **リファクタリング**: 共通処理は適切に抽象化

### デプロイ

```bash
# Cloud Buildでデプロイ
gcloud builds submit --config cloudbuild.yaml

# 直接デプロイ（開発用）
gcloud run deploy my-api \
  --source . \
  --region asia-northeast1 \
  --allow-unauthenticated
```

## 10. 監視とログ

### ログ出力例

```go
import "log"

func CreateUser(client *firestore.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        log.Printf("Creating user with IP: %s", c.ClientIP())
        
        // ... 実装 ...
        
        log.Printf("User created successfully: %s", user.ID)
    }
}
```

### Cloud Loggingでの確認

```bash
# Cloud Run のログ確認
gcloud logs read "resource.type=cloud_run_revision" --limit=50
```

## 11. パフォーマンス最適化

### Firestoreクエリ最適化

```go
// インデックスを活用したクエリ
func GetUsersByEmail(client *firestore.Client, email string) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := context.Background()
        
        iter := client.Collection("users").
            Where("email", "==", email).
            Limit(1).
            Documents(ctx)
        
        // ... 処理 ...
    }
}
```

### コネクションプーリング

```go
// メイン関数でクライアントを1回だけ初期化
// ハンドラー関数で再利用
```

## 12. セキュリティ考慮事項

### 入力値検証

```go
// Ginのバリデーション機能を活用
type CreateUserRequest struct {
    Name  string `json:"name" binding:"required,min=1,max=100"`
    Email string `json:"email" binding:"required,email"`
}
```

### 認証（必要に応じて）

```go
// JWT認証の例
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        // JWT検証ロジック
        c.Next()
    }
}
```

## まとめ

この構成により、以下が実現できます：

- ✅ **高速なコールドスタート**: Go + Cloud Run
- ✅ **充実した自動テスト**: Firestoreエミュレータ活用
- ✅ **無料枠での運用**: 低コストでの運用
- ✅ **Claude Codeでの効率的な開発**: 段階的な実装とテスト
- ✅ **スケーラブルな設計**: 将来の拡張にも対応

このガイドを参考に、段階的に機能を実装していくことをお勧めします。