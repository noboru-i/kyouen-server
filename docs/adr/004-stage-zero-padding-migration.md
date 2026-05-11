# ADR 004: ステージデータのゼロパディング補正方式

## ステータス

採用済み (2026-05-12)

## コンテキスト

旧システム（Python 時代）のバグにより、`size=6`（6×6 グリッド）のステージが `stage` フィールドに 81 文字（9×9 = 81）の文字列として保存されていた。

正しい文字列長は `size * size = 36` であるべきところ、末尾の 45 文字がすべて `"0"` のゼロパディングになっていた。

本番 Datastore（`my-android-server`）で全ステージ 12,227 件を検証した結果、以下の状況が確認された。

| カテゴリ | 件数 |
|---|---|
| 正常なステージ | 8,383 |
| ゼロパディングあり（先頭 36 文字は有効な共円を持つ） | 3,844 |
| ゼロパディングあり（先頭部分も無効） | 0 |
| その他の不正ステージ | 0 |

3,844 件はいずれも先頭 36 文字が有効な共円を持つ正しいステージデータであり、末尾のゼロパディングのみが問題であった。

このパディングはアプリ側では特に問題を起こしていなかったが、今後の重複チェック（`CheckStageExists`）や新規ステージ作成ロジックに影響する可能性があるため補正する。

## 決定事項

**専用の補正 CLI (`cmd/migrate_stage`) を 1 回限りのツールとして新設し、手動で実行する**ことにした。

### ツール構成

```
cmd/migrate_stage/
  verify/main.go    ← 全ステージを検証し、不正データを分類・JSON 出力
  migrate/main.go   ← ゼロパディングを除去して Datastore を更新
  data/
    padded_stages_before.json  ← 補正前データのバックアップ（3844 件）
```

### 補正の処理内容

1. 全 `KyouenPuzzle` エンティティを `stageNo` 順で取得
2. `len(stage) > size*size` かつ `stage[size*size:]` がすべて `"0"` のエンティティを抽出
3. `stage` フィールドを `stage[:size*size]` に切り詰めて `PutMulti`（500 件バッチ）

### dry-run モード

引数なしで dry-run として動作し、変更内容のみ標準出力に表示する。`--apply` を指定した場合のみ実際に更新を行い、更新前に `padded_stages_before.json` へバックアップを保存する。

## 検討した代替案

### 案 A: gcloud Datastore エクスポート → ローカル加工 → インポート

`gcloud datastore export` で全データを GCS にエクスポートし、ローカルで加工してインポートする。

**却下理由**: エクスポート/インポートの手順が複雑で、インポート時に既存データを上書きするリスクがある。また、対象は `stage` フィールドのみの小さな変更であり、フルエクスポートは過剰。

### 案 B: 採用案（専用 CLI）

**採用理由**:
- dry-run で実行前に補正内容を確認できる
- `PutMulti` の 500 件バッチ処理により短時間で完了する
- 補正前データを JSON ファイルとして残すことでロールバック時の参照が可能
- 既存の `DatastoreService` パターンと一致し、シンプルに実装できる

## 実行手順

### 前提

- ローカル PC に gcloud CLI がインストールされていること
- `gcloud auth application-default login` で本番プロジェクトへの認証が完了していること

### 手順

```bash
# 1. 検証（不正データの確認と補正前データの保存）
go run ./cmd/migrate_stage/verify/ cmd/migrate_stage/data/padded_stages_before.json

# 2. dry-run で補正内容を確認（書き込みは行われない）
go run ./cmd/migrate_stage/migrate/

# 3. 内容を確認し、問題なければ本実行
go run ./cmd/migrate_stage/migrate/ --apply
```

### 補正後の確認

```bash
# 再度 verify を実行し、ゼロパディング件数が 0 になっていることを確認
go run ./cmd/migrate_stage/verify/
```

## 影響

- `cmd/migrate_stage/verify/main.go`: 検証 CLI の新設
- `cmd/migrate_stage/migrate/main.go`: 補正 CLI の新設
- `cmd/migrate_stage/data/padded_stages_before.json`: 補正前データのバックアップ

## トレードオフ・注意事項

- 補正対象はすべて `size=6` のステージ（StageNo 1016 〜 12159 の範囲）であり、現行の `size=6` ステージ追加ロジックではこのバグは再現しない。
- `PutMulti` はトランザクション外のため、処理中断時に一部のレコードが補正されない可能性がある。その場合は CLI を再実行することで残余レコードを補正できる（冪等性あり）。
- バックアップ JSON は `originalStage`（81 文字）と `trimmedStage`（36 文字）の両方を保持しており、必要に応じて元の値に戻す逆補正 CLI を作成可能。
