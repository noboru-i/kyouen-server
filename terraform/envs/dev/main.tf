terraform {
  required_version = ">= 1.9"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 7.0"
    }
  }

  # リモートバックエンド（GCS）
  # 初回セットアップ時: terraform init -backend-config="bucket=<バケット名>"
  backend "gcs" {
    bucket = "api-project-732262258565-tfstate"
    prefix = "terraform/state"
  }
}

provider "google" {
  project               = var.project_id
  region                = var.region
  user_project_override = true
}

module "firebase_auth" {
  source = "../../modules/firebase-auth"

  project_id         = var.project_id
  firebase_subdomain = var.project_id

  # DEV環境: 匿名認証は無効、Twitter/Apple認証は有効
  anonymous_auth_enabled = false

  twitter_idp_enabled   = true
  twitter_client_id     = var.twitter_client_id
  twitter_client_secret = var.twitter_client_secret

  apple_idp_enabled   = true
  apple_client_id     = var.apple_client_id
  apple_client_secret = var.apple_client_secret
}

module "artifact_registry" {
  source = "../../modules/artifact-registry"

  project_id = var.project_id
  location   = var.region
}

module "cloud_run" {
  source = "../../modules/cloud-run"

  project_id            = var.project_id
  environment           = "dev"
  region                = var.region
  container_image       = var.container_image
  service_account_email = var.service_account_email
  max_instances         = 10
  min_instances         = 0

  depends_on = [module.artifact_registry]
}
