# 共円サーバー Cloud Run + Firestore 移行計画

## 概要
現在のApp Engine + Google Cloud Datastore構成から、Cloud Run + Firestore構成への移行を実施する。

## 現状分析

### 現在の構成
- **プラットフォーム**: Google App Engine (GAE)
- **データベース**: Google Cloud Datastore
- **言語**: Go
- **フレームワーク**: net/http + Gorilla Mux（推測）
- **プロジェクトID**: my-android-server

### 主要コンポーネント
- **main.go**: サーバー設定とルーティング
- **handlers/**: APIハンドラー群
  - stages_handler.go: ステージ管理
  - stages/stage_no/clear_handler.go: ステージクリア管理
  - users/login_handler.go: ユーザー認証
  - statics_handler.go: 統計情報
- **models/**: データモデル
  - kyouen.go: 共円判定ロジック
  - line.go, point.go: 幾何計算
- **db/**: データアクセス層
  - db.go: Datastore接続
  - models.go: エンティティ定義

## 移行計画

### Phase 1: 分析と準備 ✅ 実行中
- [x] 既存コードベースの詳細分析
- [x] 移行計画の作成
- [ ] 必要な依存関係の調査
- [ ] データスキーマの Firestore 互換性確認

### Phase 2: 基盤構築
- [ ] Gin フレームワークの導入
- [ ] Firestore クライアントの設定
- [ ] 新しい main.go の実装
- [ ] 基本的なルーティング設定

### Phase 3: データアクセス層の移行
- [ ] Datastore から Firestore への接続変更
- [ ] models の Firestore タグ対応
- [ ] データ操作ロジックの変換
- [ ] トランザクション処理の調整

### Phase 4: ハンドラーの移行
- [ ] stages_handler.go の Gin 対応
- [ ] users/login_handler.go の Gin 対応
- [ ] clear_handler.go の Gin 対応
- [ ] statics_handler.go の Gin 対応

### Phase 5: コンテナ化
- [ ] Dockerfile の作成
- [ ] Cloud Run 用設定
- [ ] 環境変数の整理
- [ ] ヘルスチェックエンドポイント

### Phase 6: テスト環境構築
- [ ] Firestore エミュレータ設定
- [ ] 既存テストの移行
- [ ] 新しいテストの追加
- [ ] CI/CD パイプライン更新

### Phase 7: デプロイメント
- [ ] Cloud Build 設定
- [ ] Cloud Run デプロイ設定
- [ ] 本番環境テスト
- [ ] パフォーマンス検証

## 技術的課題と対応

### 1. データモデルの互換性
**課題**: DatastoreとFirestoreのデータ形式の違い
**対応**: 
- Firestore タグの追加
- 日時フィールドの形式統一
- ID フィールドの取り扱い変更

### 2. クエリ構文の変更
**課題**: Datastore と Firestore のクエリ方法の違い
**対応**:
- Where句の書き換え
- インデックス設定の見直し
- ページネーション方法の変更

### 3. トランザクション処理
**課題**: トランザクション構文の差異
**対応**:
- Firestore のトランザクション API への移行
- エラーハンドリングの調整

### 4. 認証とセキュリティ
**課題**: OAuth設定の移行
**対応**:
- 環境変数の再設定
- セキュリティルールの追加

## 依存関係の変更

### 追加する依存関係
```bash
go get github.com/gin-gonic/gin
go get cloud.google.com/go/firestore
go get google.golang.org/api/iterator
go get github.com/stretchr/testify/assert
```

### 削除する依存関係
- Google App Engine 固有のライブラリ
- 古い Datastore クライアント

## リスク管理

### 高リスク
- **データ移行**: 既存データの保全
- **API互換性**: 既存クライアントとの互換性維持
- **OAuth設定**: 認証フローの継続性

### 対策
- ステージング環境での十分なテスト
- API仕様の詳細比較
- ロールバック計画の準備

## 成功指標

### 機能面
- [ ] 全APIエンドポイントの正常動作
- [ ] 共円判定ロジックの正確性維持
- [ ] ユーザー認証の継続性

### パフォーマンス面
- [ ] レスポンス時間の維持・改善
- [ ] コールドスタート時間の短縮
- [ ] コスト効率の改善

### 運用面
- [ ] ログ出力の適切性
- [ ] モニタリング設定
- [ ] デプロイの自動化

## 次のアクション
1. 既存コードの詳細分析
2. 依存関係の具体的調査
3. データスキーマの Firestore 対応確認