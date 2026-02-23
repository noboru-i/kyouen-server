# Terraform IaC

Firebase / GCP リソースの Infrastructure as Code（Terraform）定義です。

## ディレクトリ構成

```
terraform/
├── README.md
├── modules/
│   ├── kyouen-app/          # 共通アプリケーションモジュール（各環境から呼び出す）
│   ├── firebase-auth/       # Firebase Authentication モジュール
│   ├── cloud-run/           # Cloud Run サービス モジュール
│   └── artifact-registry/   # Artifact Registry モジュール
└── envs/
    ├── dev/                 # DEV環境 (api-project-732262258565)
    └── prod/                # PROD環境 (my-android-server)
```

## 管理対象リソース

| リソース | Terraform リソース | DEV | PROD |
|----------|-------------------|-----|------|
| Firebase Auth 基本設定 | `google_identity_platform_config` | ✅ | ✅ |
| Twitter IdP | `google_identity_platform_default_supported_idp_config` | ✅ | - |
| Apple IdP | `google_identity_platform_default_supported_idp_config` | ✅ | - |
| Cloud Run サービス | `google_cloud_run_v2_service` | ✅ | ✅ |
| Cloud Run IAM（公開） | `google_cloud_run_v2_service_iam_member` | ✅ | ✅ |
| Artifact Registry | `google_artifact_registry_repository` | ✅ | ✅ |

> **Firestore（Datastoreモード）は管理対象外**: 既存データが入っており削除リスクがあるため。設定変更が必要な場合は手動操作またはgcloud CLIで実施。

## セットアップ

### 前提条件

```bash
# Terraform インストール
brew install terraform

# GCloud 認証
gcloud auth application-default login
```

### 初回セットアップ（既存リソースのインポート）

既存のリソースを Terraform 管理下に置くため、インポートが必要です。

```bash
# DEV環境
cd terraform/envs/dev
terraform init

# DEV Firebase Auth
terraform import module.kyouen_app.module.firebase_auth.google_identity_platform_config.auth projects/api-project-732262258565/config

# DEV Twitter IdP
terraform import 'module.kyouen_app.module.firebase_auth.google_identity_platform_default_supported_idp_config.twitter[0]' \
  projects/api-project-732262258565/defaultSupportedIdpConfigs/twitter.com

# DEV Apple IdP
terraform import 'module.kyouen_app.module.firebase_auth.google_identity_platform_default_supported_idp_config.apple[0]' \
  projects/api-project-732262258565/defaultSupportedIdpConfigs/apple.com

# DEV Artifact Registry
terraform import module.kyouen_app.module.artifact_registry.google_artifact_registry_repository.kyouen_repo \
  projects/api-project-732262258565/locations/asia-northeast1/repositories/kyouen-repo

# DEV Cloud Run
terraform import module.kyouen_app.module.cloud_run.google_cloud_run_v2_service.kyouen_server \
  projects/api-project-732262258565/locations/asia-northeast1/services/kyouen-server-dev

# DEV Cloud Run IAM
terraform import module.kyouen_app.module.cloud_run.google_cloud_run_v2_service_iam_member.public_access \
  "projects/api-project-732262258565/locations/asia-northeast1/services/kyouen-server-dev roles/run.invoker allUsers"
```

```bash
# PROD環境
cd terraform/envs/prod
terraform init

# PROD Firebase Auth
terraform import module.kyouen_app.module.firebase_auth.google_identity_platform_config.auth projects/my-android-server/config

# PROD Artifact Registry
terraform import module.kyouen_app.module.artifact_registry.google_artifact_registry_repository.kyouen_repo \
  projects/my-android-server/locations/asia-northeast1/repositories/kyouen-repo

# PROD Cloud Run
terraform import module.kyouen_app.module.cloud_run.google_cloud_run_v2_service.kyouen_server \
  projects/my-android-server/locations/asia-northeast1/services/kyouen-server-prod

# PROD Cloud Run IAM
terraform import module.kyouen_app.module.cloud_run.google_cloud_run_v2_service_iam_member.public_access \
  "projects/my-android-server/locations/asia-northeast1/services/kyouen-server-prod roles/run.invoker allUsers"
```

## 通常の使い方

```bash
# プラン確認
cd terraform/envs/dev
terraform plan -var="twitter_client_id=..." -var="twitter_client_secret=..." \
               -var="apple_client_id=..." -var="apple_client_secret=..."

# 適用
terraform apply

# 機密変数はファイルで管理する場合（.gitignore に追加済み）
cp terraform.tfvars.example terraform.tfvars
# terraform.tfvars を編集して機密値を設定
terraform apply -var-file="terraform.tfvars"
```

## 機密情報の管理

以下の変数は機密情報のため、以下のいずれかで管理してください：

- **GitHub Actions**: GitHub Secrets として設定し、CI/CD 経由で `TF_VAR_xxx` として渡す
- **ローカル開発**: `terraform.tfvars`（gitignore 対象）に記載
- **Secret Manager**: GCP Secret Manager から参照する（将来的な改善案）

機密変数一覧（DEV）:
- `twitter_client_id`
- `twitter_client_secret`
- `apple_client_id`
- `apple_client_secret`

## 注意事項

1. **Firestore は管理外**: データが入っているため、誤ってデストロイしないよう意図的に管理対象外としています
2. **container_image**: デプロイごとに変わるため、`terraform apply` では通常変更不要。CI/CD（GitHub Actions）が `gcloud run deploy` で直接更新します
3. **Apple IdP の privateKey**: `google_identity_platform_default_supported_idp_config` リソースでは Apple の `appleSignInConfig`（teamId, keyId, privateKey）を設定できません。Firebase コンソールまたは REST API で手動設定が必要です
