# DEV / PROD 環境設定差異

DEV環境と本番（PROD）環境の設定差異をまとめます。

## Firebase / GCP プロジェクトID

| 項目 | DEV | PROD |
|------|-----|------|
| Project ID | `api-project-732262258565` | `my-android-server` |

## GitHub Actions Secrets

| シークレット | DEV | PROD |
|------|-----|------|
| WIF Provider | `DEV_WIF_PROVIDER` | `PROD_WIF_PROVIDER` |
| WIF Service Account | `DEV_WIF_SERVICE_ACCOUNT` | `PROD_WIF_SERVICE_ACCOUNT` |

## Cloud Run サービス名・環境変数

| 項目 | DEV | PROD |
|------|-----|------|
| Cloud Run サービス名 | `kyouen-server-dev` | `kyouen-server-prod` |
| `ENVIRONMENT` 環境変数 | `dev` | `prod` |
| `GOOGLE_CLOUD_PROJECT` 環境変数 | `api-project-732262258565` | `my-android-server` |

## デプロイトリガー

| 項目 | DEV | PROD |
|------|-----|------|
| トリガー | `main` ブランチへのプッシュ（自動） | `workflow_dispatch`（手動・確認入力必須） |
| ワークフロー | `.github/workflows/deploy-dev.yml` | `.github/workflows/deploy-prod.yml` |

## Firebase プロジェクト切り替え

`.firebaserc` に `dev` / `prod` エイリアスが設定されており、以下のコマンドで切り替えられます。

```bash
firebase use dev   # DEV環境 (api-project-732262258565)
firebase use prod  # 本番環境 (my-android-server)
```

## 関連ファイル

- [`.firebaserc`](../.firebaserc) - Firebase プロジェクトエイリアス定義
- [`internal/config/config.go`](../internal/config/config.go) - `ENVIRONMENT` による ProjectID の切り替えロジック
- [`.github/workflows/deploy-dev.yml`](../.github/workflows/deploy-dev.yml) - DEV デプロイワークフロー
- [`.github/workflows/deploy-prod.yml`](../.github/workflows/deploy-prod.yml) - PROD デプロイワークフロー
- [`.github/workflows/deploy-common.yml`](../.github/workflows/deploy-common.yml) - 共通デプロイロジック
