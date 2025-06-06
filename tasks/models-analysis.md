# モデル詳細分析

## データベースモデル (`db/models.go`)

### 現在のDatastoreエンティティ

#### 1. KyouenPuzzleSummary
```go
type KyouenPuzzleSummary struct {
    Count    int64     `datastore:"count"`
    LastDate time.Time `datastore:"lastDate"`
}
```
**用途**: パズル統計情報の管理
**Firestore移行**: 最小限の変更

#### 2. KyouenPuzzle
```go
type KyouenPuzzle struct {
    StageNo    int64     `datastore:"stageNo"`
    Size       int64     `datastore:"size"`
    Stage      string    `datastore:"stage"`
    Creator    string    `datastore:"creator"`
    RegistDate time.Time `datastore:"registDate"`
}
```
**用途**: パズルステージの管理
**特徴**: StageNo による順序付け、Stage 文字列による盤面表現

#### 3. User
```go
type User struct {
    UserID          string `datastore:"userId"`
    ScreenName      string `datastore:"screenName"`
    Image           string `datastore:"image"`
    ClearStageCount int64  `datastore:"clearStageCount"`
    
    // レガシーフィールド（削除予定）
    AccessToken  string `datastore:"accessToken"`
    AccessSecret string `datastore:"accessSecret"`
    APIToken     string `datastore:"apiToken"`
}
```
**用途**: ユーザー情報の管理
**課題**: レガシーフィールドの整理が必要

#### 4. StageUser
```go
type StageUser struct {
    StageKey  *datastore.Key `datastore:"stage"`
    UserKey   *datastore.Key `datastore:"user"`
    ClearDate time.Time      `datastore:"clearDate"`
}
```
**用途**: ユーザーのステージクリア記録
**課題**: Datastore Key参照の Firestore 対応

## ビジネスロジックモデル (`models/`)

### 共円判定の核心ロジック

#### 1. KyouenStage (`models/kyouen.go`)
```go
type KyouenStage struct {
    size                int      // グリッドサイズ
    stonePointList      []Point  // 黒石の位置（パズル石）
    whiteStonePointList []Point  // 白石の位置（ユーザー解答）
}
```

**主要メソッド**:
- `NewKyouenStage(size int, stage string)`: 文字列からステージ作成
- `NewRotatedKyouenStage()`: 90度回転
- `NewMirroredKyouenStage()`: 鏡像反転
- `HasKyouen()`: 共円判定（黒石）
- `IsKyouenByWhite()`: 共円判定（白石）

#### 2. KyouenData
```go
type KyouenData struct {
    points     []Point      // 共円を構成する点
    lineKyouen bool         // 直線かどうか
    center     FloatPoint   // 円の中心
    radius     float64      // 円の半径
    line       Line         // 直線の場合のライン
}
```

**用途**: 共円判定結果の格納

#### 3. 幾何計算モデル

##### Point (`models/point.go`)
```go
type Point struct {
    x int
    y int
}

type FloatPoint struct {
    x float64
    y float64
}
```

##### Line (`models/line.go`)
```go
type Line struct {
    p1 FloatPoint
    p2 FloatPoint
    a  float64  // 直線方程式 ax + by + c = 0
    b  float64
    c  float64
}
```

## Firestore移行時の変更点

### 1. タグの変更
```go
// 現在
type KyouenPuzzle struct {
    StageNo    int64     `datastore:"stageNo"`
    Size       int64     `datastore:"size"`
    Stage      string    `datastore:"stage"`
    Creator    string    `datastore:"creator"`
    RegistDate time.Time `datastore:"registDate"`
}

// 移行後
type KyouenPuzzle struct {
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

### 2. 参照の処理
```go
// 現在（Datastore Key）
type StageUser struct {
    StageKey  *datastore.Key `datastore:"stage"`
    UserKey   *datastore.Key `datastore:"user"`
    ClearDate time.Time      `datastore:"clearDate"`
}

// 移行後（文字列ID参照）
type StageUser struct {
    ID        string    `firestore:"-" json:"id"`
    StageID   string    `firestore:"stageId" json:"stageId"`
    UserID    string    `firestore:"userId" json:"userId"`
    ClearDate time.Time `firestore:"clearDate" json:"clearDate"`
    CreatedAt time.Time `firestore:"createdAt" json:"createdAt"`
    UpdatedAt time.Time `firestore:"updatedAt" json:"updatedAt"`
}
```

### 3. コレクション設計

#### Firestore コレクション構造
```
kyouen-server/
├── users/                    # ユーザー情報
│   └── {userID}/
├── stages/                   # ステージ情報
│   └── {stageID}/
├── stage_users/              # クリア記録
│   └── {stageUserID}/
└── summaries/                # 統計情報
    └── puzzle_summary
```

## 移行作業項目

### 1. データベースモデルの更新
- [ ] Firestore タグの追加
- [ ] ID フィールドの追加
- [ ] CreatedAt/UpdatedAt フィールドの追加
- [ ] Key 参照から文字列ID参照への変更

### 2. ビジネスロジックの保持
- [ ] 共円判定ロジックは変更なし
- [ ] 幾何計算ロジックは変更なし
- [ ] ステージ回転・反転ロジックは変更なし

### 3. バリデーション追加
```go
type CreateStageRequest struct {
    Size    int64  `json:"size" binding:"required,min=3,max=10"`
    Stage   string `json:"stage" binding:"required"`
    Creator string `json:"creator" binding:"required,min=1,max=50"`
}
```

### 4. レスポンスモデル
```go
type StageResponse struct {
    ID         string    `json:"id"`
    StageNo    int64     `json:"stageNo"`
    Size       int64     `json:"size"`
    Stage      string    `json:"stage"`
    Creator    string    `json:"creator"`
    RegistDate time.Time `json:"registDate"`
}
```

## テスト戦略

### 1. ビジネスロジックテスト
- 共円判定ロジックの既存テストを維持
- 幾何計算の精度テスト
- ステージ回転・反転の正確性テスト

### 2. データアクセステスト
- Firestore エミュレータでの CRUD テスト
- クエリ性能テスト
- トランザクション整合性テスト

### 3. 統合テスト
- エンドツーエンドのステージ作成フロー
- ユーザークリア記録フロー
- 統計情報更新フロー

## パフォーマンス考慮事項

### 1. インデックス設計
```
Collection: stages
- stageNo (ASC) - ページネーション用
- creator (ASC) - 作成者別検索用

Collection: stage_users  
- userId (ASC) - ユーザー別クリア記録
- stageId (ASC) - ステージ別クリア記録
- clearDate (DESC) - 最新クリア順
```

### 2. データサイズ最適化
- Stage 文字列の効率的な格納
- 不要なフィールドの削除（レガシーOAuth情報）

### 3. クエリ最適化
- 複合インデックスの適切な設計
- ページネーションの効率化