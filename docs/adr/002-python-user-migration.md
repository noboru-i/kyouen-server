# ADR 002: kyouen-python からの User レコードマイグレーション方式

## ステータス

採用済み (2026-04-11)

## コンテキスト

kyouen-python (旧 Google App Engine / Python アプリ) から kyouen-server (Go / Cloud Run) への移行において、DatastoreのUserエンティティのキー構造が異なる問題が発生した。

| システム | キー形式 | userId フィールド |
|---|---|---|
| kyouen-python | `'KEY' + Twitter UID` | Twitter UID |
| kyouen-server | `'KEY' + Firebase UID` | Firebase UID |

Firebase Auth 経由の Twitter ログインでは、Firebase UID と Twitter UID は異なる値となる。
そのため、Go側がログイン処理を行う際に既存の Python 時代のエンティティを発見できず、新規ユーザーとして作成してしまい、`clearStageCount` や `StageUser` レコードが失われるという問題があった。

## 決定事項

**初回ログイン時に自動マイグレーションを実行する** ことにした。

`CreateOrUpdateUserFromFirebase` 関数において、Firebase UID でユーザーが見つからない場合、Twitter UID による追加検索を行い、レガシーエンティティが見つかれば透過的にマイグレーションする。

### マイグレーションの処理内容

トランザクション内（原子性を保証）:
1. 旧エンティティ (`'KEY' + twitterUID`) を読み取り、`clearStageCount` を取得
2. 新エンティティ (`'KEY' + firebaseUID`) を作成（`clearStageCount` 引き継ぎ）
3. `UserMigration` レコードを作成（監査ログ）
4. 旧エンティティを削除

トランザクション外（Datastore の 25 エンティティグループ制限のため）:
5. 旧ユーザーキーを参照する `StageUser` レコードを新キーに更新

## 検討した代替案

### 案 A: バッチマイグレーションスクリプト

デプロイ前に全 User エンティティを一括マイグレーションする。

**却下理由**: Firebase UID と Twitter UID の対応関係を事前に取得するためには全ユーザーを Firebase Auth に問い合わせる必要があり、大量の API 呼び出しが発生する。また、実際にログインしないユーザーのデータまで移行する必要はない。

### 案 B: 両キーを並行サポート

Firebase UID と Twitter UID の両方でユーザーを検索し、両方のエンティティを保持する。

**却下理由**: データの二重管理が発生し、`clearStageCount` の整合性維持が複雑になる。

### 案 C: 採用案（初回ログイン時の自動マイグレーション）

**採用理由**:
- 実際にログインするユーザーのみマイグレーション対象となり効率的
- ユーザー側の操作不要で透過的に行われる
- `UserMigration` レコードによりマイグレーション状況の監査が可能

## 影響

- `internal/datastore/datastore.go`: `MigrateLegacyUser`、`migrateStageUserRecords` の追加、`CreateOrUpdateUserFromFirebase` の修正
- `internal/datastore/models.go`: `UserMigration` 構造体の追加
- `docs/datastore-schema.json`: `UserMigration` Kind の定義追加
- `docs/auth-data-relations.md`: マイグレーションフローの説明追加

## トレードオフ・注意事項

- `StageUser` レコードの更新はトランザクション外のため、処理中断時に一部のレコードが旧キーを参照したままになる可能性がある。ただし `clearStageCount` はトランザクション内で正しく引き継がれており、実質的なデータロスは発生しない。`UserMigration` レコードを参照することで、後から旧キーを特定して再修復も可能。
- `twitterUID` が空（将来的な非 Twitter プロバイダ）の場合は、レガシー検索をスキップして通常の新規作成フローに入る。
