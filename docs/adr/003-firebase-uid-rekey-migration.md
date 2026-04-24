# ADR 003: Firebase UID 再発行に伴うユーザーデータ補正方式

## ステータス

採用済み (2026-04-24)

## コンテキスト

本番 Firebase Authentication 上のユーザーアカウントをオペレーションミスで削除してしまい、同じ Twitter アカウントで再作成したところ Firebase UID が変わってしまった。

Datastore 上の `User` エンティティはキー名が `"KEY" + Firebase UID` となっており、`StageUser` エンティティもこの User キーを `user` プロパティとして参照している。UID が変わると以下のデータが失われる。

| エンティティ / フィールド | 問題 |
|---|---|
| `User` エンティティキー | 旧 UID のキーが残り、新 UID でのログイン時に新規ユーザーとして作成される |
| `User.clearStageCount` | 旧ユーザーのクリア数が引き継がれない |
| `StageUser.user` | 旧 User キーを参照したまま、新 UID 側では参照されない |

また、アカウント再作成後にアプリからログインを行ったため、新 UID の `User` エンティティが `clearStageCount=0` の状態で自動生成されていた。

Firebase Auth の「UID の変更」という操作は存在しないため、Datastore 側のデータを手動で補正する必要がある。

## 決定事項

**専用の補正 CLI (`cmd/migrate_user`) を 1 回限りのツールとして新設し、手動で実行する**ことにした。

既存の ADR 002 で実装済みの `MigrateLegacyUser`（Python 時代の Twitter UID キー → Firebase UID キーの移行）と同一のデータ処理パターン（トランザクション内での User 差し替え + トランザクション外での StageUser 再ポイント）を流用する。

### 補正の処理内容

事前検証:
- 旧 UID / 新 UID の `User` エンティティがどちらも存在することを確認
- `twitterUid` フィールドの一致で同一人物であることを確認（誤補正防止）
- 新 UID 側の `StageUser` が 0 件であることを確認（ありえないが念のため）

トランザクション内（原子性を保証）:
1. 旧 User エンティティを読み取り、`clearStageCount` を取得
2. 新 User エンティティを読み取り、`clearStageCount` を旧ユーザー値で上書きして `Put`（`screenName` / `image` / `twitterUid` は新 User 側の値を維持）
3. `UserMigration` レコードを作成（監査ログ。`OldKey` に旧UID、`NewKey` に新UID、`FirebaseUID` に新UID を記録）
4. 旧 User エンティティを削除

トランザクション外（Datastore の 25 エンティティグループ制限のため）:
5. 旧 User キーを参照する `StageUser` レコードを新 User キーに更新

### dry-run モード

`-dry-run` フラグで書き込みなしの事前確認が可能。本実行前に必ず dry-run を実施する。

## 検討した代替案

### 案 A: 通常ログインフローでの自動補正

`CreateOrUpdateUserFromFirebase` にロジックを追加し、「旧 UID のユーザーが存在し twitterUid が一致する場合は自動マイグレーション」を行う。

**却下理由**: 通常ログインフローへの影響範囲が大きく、1 回限りの補正のためにランタイムに常時ロジックを残すのは好ましくない。また、今回は新 UID 側にも既に User エンティティが存在するため、既存の `CreateOrUpdateUserFromFirebase` の分岐（`ErrNoSuchEntity` のみ）には乗れない。

### 案 B: gcloud / Datastore コンソールで手動操作

コンソールからエンティティのキーを書き換える。

**却下理由**: Datastore ではキー名の変更ができないため、削除→再作成が必要になる。`StageUser` の書き換えも手作業では件数が多く、ミスのリスクが高い。

### 案 C: 採用案（専用 CLI）

**採用理由**:
- 既存の `MigrateLegacyUser` / `migrateStageUserRecords` の実績ある実装を流用でき、安全
- dry-run で実行前に内容を確認できる
- `twitterUid` 一致チェックにより誤補正を防止できる
- 事後に件数を照合して補正の完全性を確認できる
- 1 回限り利用後は削除 or 残置を選択できる

## 実行手順

### 前提

- ローカル PC に gcloud CLI がインストールされていること
- `go run` でビルド可能な環境であること
- 旧 Firebase UID と新 Firebase UID を把握していること

### 手順

```bash
# 1. Application Default Credentials で本番 Datastore に接続できるよう認証
gcloud auth application-default login

# 2. dry-run で補正内容を確認（書き込みは行われない）
GOOGLE_CLOUD_PROJECT=my-android-server \
  go run cmd/migrate_user/main.go \
  -old-uid=<旧FirebaseUID> \
  -new-uid=<新FirebaseUID> \
  -dry-run

# dry-run の出力例:
# 旧ユーザー: screenName=noboru, twitterUid=12345678, clearStageCount=42
# 新ユーザー: screenName=noboru, twitterUid=12345678, clearStageCount=0
# twitterUid 一致確認: OK (12345678)
# 旧ユーザー StageUser 件数: 42 件
# 新ユーザー StageUser 件数: 0 件
# 補正内容:
#   旧 User エンティティ (KEY<旧UID>) を削除
#   新 User エンティティ (KEY<新UID>) の clearStageCount を 42 に更新
#   StageUser 42 件の UserKey を新 UID に差し替え
#   UserMigration レコードを 1 件追加
# dry-run 完了。上記の変更は行われていません。

# 3. 内容を確認し、問題なければ本実行
GOOGLE_CLOUD_PROJECT=my-android-server \
  go run cmd/migrate_user/main.go \
  -old-uid=<旧FirebaseUID> \
  -new-uid=<新FirebaseUID>
```

### 補正後の確認

1. `UserMigration` エンティティに本日付の新規レコードが 1 件追加されていること
2. Datastore コンソールで `KEY<旧UID>` の User が存在しないこと
3. `KEY<新UID>` の User の `clearStageCount` が補正前の値になっていること
4. アプリから新 UID でログインし `GET /v2/stages` でクリア済みステージが正しく返ること
5. CLI の事後確認出力で旧ユーザー StageUser 件数が 0、新ユーザー StageUser 件数が期待値と一致していること

## 影響

- `cmd/migrate_user/main.go`: 補正 CLI の新設
- `internal/datastore/datastore.go`: `MigrateFirebaseUID`、`CountStageUsersByUserKey` の追加

## トレードオフ・注意事項

- ADR 002 と同様に、`StageUser` レコードの更新はトランザクション外のため、処理中断時に一部のレコードが旧キーを参照したままになる可能性がある。CLI を再実行することで残余レコードを補正できる（残存した旧 User エンティティが存在しないため再実行は旧キーの Get が失敗してトランザクションがエラーになるが、StageUser 再ポイントは `migrateStageUserRecords` を直接呼ぶ手順で個別対応が可能）。
- `UserMigration.TwitterUID` フィールドには旧 Firebase UID ではなく Twitter UID が入る（既存スキーマの制約）。`OldKey` / `NewKey` / `MigratedAt` で補正の追跡は十分可能。
- 新 UID 側に `StageUser` が存在する場合は安全のため CLI が中断する。その場合は StageUser のマージ戦略を別途検討すること。
