# 次のステップ: Cloud Run + Firestore 移行

## 📋 完了した作業

### ✅ 分析フェーズ
- [x] 既存コードベースの詳細分析
- [x] ハンドラーの構造と依存関係の把握
- [x] データモデルとビジネスロジックの理解
- [x] 移行計画書の作成
- [x] 実装計画の策定

### 📁 作成したドキュメント
1. `tasks/migration-plan.md` - 移行の概要と戦略
2. `tasks/current-analysis.md` - 現状のアーキテクチャ分析
3. `tasks/handlers-analysis.md` - ハンドラーの詳細分析
4. `tasks/models-analysis.md` - データモデルの詳細分析
5. `tasks/implementation-plan.md` - 段階的な実装計画

## 🚀 次に実行すべき作業

### Phase 1: 基盤構築（即座に開始可能）

#### 1.1 依存関係の更新
```bash
# 新しい依存関係を追加
go get github.com/gin-gonic/gin@latest
go get cloud.google.com/go/firestore@latest
go get google.golang.org/api/iterator@latest
go get github.com/stretchr/testify/assert@latest
go get github.com/gin-contrib/cors@latest

# go.modの整理
go mod tidy
```

#### 1.2 基本ディレクトリ構造の作成
```bash
mkdir -p {config,services,middleware,handlers/v2,models/firestore,testutils}
```

#### 1.3 設定管理の実装
**ファイル**: `config/config.go`
- 環境変数の統一管理
- 設定値のバリデーション
- 開発/本番環境の切り替え

### Phase 2: Firestore基盤（1-2日）

#### 2.1 Firestoreサービス層の作成
**ファイル**: `services/firestore.go`
- Firestoreクライアントの初期化
- 基本的なCRUD操作の実装
- エラーハンドリングの統一

#### 2.2 新しいmain.goの並行開発
**ファイル**: `main_new.go`
- Gin エンジンの設定
- ミドルウェアの設定
- ルーティングの実装

### Phase 3: 最初のAPIエンドポイント（1-2日）

#### 3.1 StaticsHandler の移行
**理由**: 最もシンプルで依存関係が少ない
**ファイル**: `handlers/v2/statics.go`

#### 3.2 ヘルスチェックエンドポイント
**ファイル**: `handlers/v2/health.go`
- `/health` エンドポイント
- Firestore接続確認

## 🛠️ 実装の開始手順

### Step 1: 作業ブランチの作成
```bash
git checkout -b feature/cloudrun-firestore-migration
```

### Step 2: 依存関係の更新
```bash
# 実行コマンド
go get github.com/gin-gonic/gin@latest
go mod tidy
```

### Step 3: 基本構造の作成
```bash
# ディレクトリ作成
mkdir -p {config,services,middleware,handlers/v2,models/firestore,testutils}

# 設定ファイルの作成
touch config/config.go
touch services/firestore.go
touch main_new.go
```

### Step 4: 環境設定
```bash
# .env.example ファイルの作成
cat > .env.example << 'EOF'
PORT=8080
GOOGLE_CLOUD_PROJECT=my-android-server
FIRESTORE_EMULATOR_HOST=localhost:8080
CONSUMER_KEY=your_twitter_consumer_key
CONSUMER_SECRET=your_twitter_consumer_secret
FIREBASE_CREDENTIALS_FILE=path/to/service-account.json
GIN_MODE=debug
EOF
```

## 🔄 開発フロー

### 1. 段階的な実装
- 1つのエンドポイントずつ移行
- 既存APIとの並行運用
- 段階的なテスト実行

### 2. テスト戦略
```bash
# Firestoreエミュレータ起動
firebase emulators:start --only firestore

# テスト実行
go test ./... -v

# 統合テスト
# 既存APIと新APIの結果比較
```

### 3. デプロイ準備
- ローカル環境での動作確認
- Dockerコンテナでの動作確認
- Cloud Runでのテストデプロイ

## 📊 進捗管理

### マイルストーン
1. **Week 1**: 基盤構築とStaticsAPI移行
2. **Week 2**: StagesAPI（GET）の移行
3. **Week 3**: StagesAPI（POST）と認証の移行
4. **Week 4**: ClearAPIとテスト完了
5. **Week 5**: 本番移行とモニタリング

### 成功指標
- [ ] 全APIエンドポイントの動作確認
- [ ] 既存データの完全移行
- [ ] パフォーマンスの維持・改善
- [ ] レスポンス時間の測定
- [ ] エラー率の監視

## ⚠️ 注意事項

### 並行開発
- 既存システムを停止させない
- `main_new.go` での並行開発
- 新旧APIの共存期間を設ける

### データ整合性
- 移行前の完全バックアップ
- 段階的なデータ移行
- ロールバック計画の準備

### セキュリティ
- サービスアカウントキーの適切な管理
- 環境変数での秘匿情報管理
- ファイアウォールルールの設定

## 🔗 参考資料

### 技術ドキュメント
- [Gin Framework Documentation](https://gin-gonic.com/docs/)
- [Cloud Firestore Go Client](https://cloud.google.com/firestore/docs/reference/libraries#client-libraries-install-go)
- [Cloud Run Go Quickstart](https://cloud.google.com/run/docs/quickstarts/build-and-deploy/deploy-go-service)

### 既存資料
- `docs/guide.md` - Cloud Run + Firestore 開発ガイド
- `CLAUDE.md` - プロジェクト固有の指示

## 💬 サポート

### 質問・相談
- 実装中の技術的な質問
- アーキテクチャの判断が必要な場合
- パフォーマンスの最適化

### 継続的な支援
- 各フェーズでの進捗確認
- 問題発生時のトラブルシューティング
- コードレビューとベストプラクティス提案

---

**次のアクション**: Phase 1の依存関係更新から開始することをお勧めします。準備ができましたら、具体的な実装を始めましょう！