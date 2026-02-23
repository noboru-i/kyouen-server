terraform {
  required_version = ">= 1.9"

  required_providers {
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 7.0"
    }
  }

  # リモートバックエンド（GCS）
  # 初回セットアップ時: terraform init -backend-config="bucket=<バケット名>"
  backend "gcs" {
    bucket = "my-android-server-tfstate"
    prefix = "terraform/state"
  }
}

provider "google-beta" {
  project               = var.project_id
  region                = var.region
  user_project_override = true
}

module "firebase_auth" {
  source = "../../modules/firebase-auth"

  project_id         = var.project_id
  firebase_subdomain = var.project_id

  # PROD環境: 匿名認証は有効、Twitter/Apple認証は現在未設定
  # TODO: PRODへのIdP展開が必要か要確認（docs/firebase-config-comparison.md 参照）
  anonymous_auth_enabled = true

  twitter_idp_enabled = false
  apple_idp_enabled   = false
}

module "artifact_registry" {
  source = "../../modules/artifact-registry"

  project_id = var.project_id
  location   = var.region
}

module "cloud_run" {
  source = "../../modules/cloud-run"

  project_id            = var.project_id
  environment           = "prod"
  region                = var.region
  container_image       = var.container_image
  service_account_email = var.service_account_email
  max_instances         = 10
  min_instances         = 0

  depends_on = [module.artifact_registry]
}
