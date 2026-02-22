# ADR-001: Terraform による Infrastructure as Code の導入

- **日付**: 2026-02-23
- **ステータス**: 承認済み

## コンテキスト

Firebase Auth・Cloud Run・Artifact Registry などの GCP/Firebase リソースが手動で管理されており、設定の再現性・変更履歴の追跡が困難だった。

## 決定事項

### 1. Terraform で管理するリソース

| リソース | Terraform リソース |
|----------|-------------------|
| Firebase Auth 基本設定 | `google_identity_platform_config` |
| Twitter / Apple IdP | `google_identity_platform_default_supported_idp_config` |
| Cloud Run サービス | `google_cloud_run_v2_service` |
| Cloud Run IAM（公開） | `google_cloud_run_v2_service_iam_member` |
| Artifact Registry | `google_artifact_registry_repository` |

**管理対象外**: Firestore（Datastoreモード）は既存データが入っているため、誤削除リスクを避けて対象外とする。

### 2. プロバイダー設定

`hashicorp/google` プロバイダー（`~> 6.0`）を使用し、**`user_project_override = true`** を必須とする。

```hcl
provider "google" {
  project               = var.project_id
  region                = var.region
  user_project_override = true
}
```

`user_project_override = true` が必要な理由: `identitytoolkit.googleapis.com`（Firebase Auth）は ADC（Application Default Credentials）使用時にクォータプロジェクトの明示が必要なため。設定なしでは Cloud SDK のデフォルトプロジェクトが使われ 403 エラーになる。

### 3. Terraform State の保存先

GCS バケット（`<project-id>-tfstate`）をリモートバックエンドとして使用する。

```hcl
backend "gcs" {
  bucket = "api-project-732262258565-tfstate"
  prefix = "terraform/state"
}
```

### 4. Container Image は Terraform で管理しない

Cloud Run の `template[0].containers[0].image` は CI/CD（GitHub Actions + `gcloud run deploy`）が管理するため、`lifecycle.ignore_changes` で除外する。

```hcl
lifecycle {
  ignore_changes = [
    template[0].containers[0].image,
    client,
    client_version,
  ]
}
```

### 5. 機密情報の管理

Twitter / Apple の OAuth 認証情報は `terraform.tfvars`（`.gitignore` 対象）で管理する。

```
terraform/envs/dev/terraform.tfvars  # gitignore 対象
```

Apple の `client_secret` は静的な文字列ではなく **ES256 JWT**（有効期限 〜6ヶ月）であり、期限切れ前に再生成・更新が必要。

### 6. 初回セットアップ手順

既存リソースは `terraform import` で取り込む（`terraform/README.md` 参照）。

## 結果

- GCP リソースの設定変更が Pull Request + `terraform plan` でレビュー可能になる
- DEV / PROD 環境の設定差分が `terraform/envs/` ディレクトリで明示的に管理される
- Apple JWT の定期更新（〜6ヶ月ごと）が運用タスクとして必要になる
