---
allowed-tools: Bash(uv run .claude/scripts/gen_apple_client_secret.py:*), Read, Edit
description: Apple Sign In 用の client_secret JWT を生成する
---

## タスク

Apple Sign In の `client_secret`（JWT）を生成して `terraform/envs/$ENV/terraform.tfvars` に書き込む。

## 事前確認

まず以下のファイルを読んで現在の設定値を確認する：
- `terraform/envs/prod/terraform.tfvars`（または指定された環境）

確認すべき値：
- `apple_client_id`（= Service ID、JWT の `sub` クレームに使う）
- 既存の `apple_client_secret` の JWT ヘッダー部分をデコードして `kid`（Key ID）を確認

## ユーザーへの確認

生成前に以下を確認する：

1. **環境**: dev / prod のどちら？
2. **`.p8` ファイルのパス**: Apple Developer Portal からダウンロードした秘密鍵ファイル（例: `~/Downloads/AuthKey_XXXXXXXXXX.p8`）
3. **Team ID**: Apple Developer Portal のメンバーシップ画面に表示される 10 桁の英数字

## JWT 生成

以下のコマンドで生成する：

```bash
uv run .claude/scripts/gen_apple_client_secret.py \
    --p8 <path_to_AuthKey.p8> \
    --team-id <TEAM_ID> \
    --key-id <KEY_ID> \
    --client-id <apple_client_id の値>
```

出力の `=== JWT Payload (verification) ===` セクションで `sub` が `apple_client_id` と一致していることを確認する。

## tfvars への書き込み

生成した JWT（`=== Generated client_secret ===` の値）を
`terraform/envs/$ENV/terraform.tfvars` の `apple_client_secret` の値に Edit ツールで更新する。
