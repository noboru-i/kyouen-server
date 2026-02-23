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
    bucket = "my-android-server-tfstate"
    prefix = "terraform/state"
  }
}

provider "google" {
  project               = var.project_id
  region                = var.region
  user_project_override = true
}

module "kyouen_app" {
  source = "../../modules/kyouen-app"

  project_id            = var.project_id
  environment           = var.environment
  region                = var.region
  container_image       = var.container_image
  service_account_email = var.service_account_email
  twitter_client_id     = var.twitter_client_id
  twitter_client_secret = var.twitter_client_secret
  apple_client_id       = var.apple_client_id
  apple_client_secret   = var.apple_client_secret
}
