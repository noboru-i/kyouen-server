terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
  }
}

# Firebase Authentication 基本設定
resource "google_identity_platform_config" "auth" {
  project = var.project_id

  sign_in {
    anonymous {
      enabled = var.anonymous_auth_enabled
    }
    email {
      enabled           = false
      password_required = false
    }
    allow_duplicate_emails = false
  }

  authorized_domains = concat(
    [
      "localhost",
      "${var.firebase_subdomain}.firebaseapp.com",
      "${var.firebase_subdomain}.web.app",
    ],
    var.additional_authorized_domains
  )

  mfa {
    state = "DISABLED"
  }

  lifecycle {
    ignore_changes = []
  }
}

# Twitter IdP設定（有効な場合のみ作成）
resource "google_identity_platform_default_supported_idp_config" "twitter" {
  count = var.twitter_idp_enabled ? 1 : 0

  project       = var.project_id
  idp_id        = "twitter.com"
  client_id     = var.twitter_client_id
  client_secret = var.twitter_client_secret
  enabled       = true
}

# Apple IdP設定（有効な場合のみ作成）
resource "google_identity_platform_default_supported_idp_config" "apple" {
  count = var.apple_idp_enabled ? 1 : 0

  project       = var.project_id
  idp_id        = "apple.com"
  client_id     = var.apple_client_id
  client_secret = var.apple_client_secret
  enabled       = true

  # Apple固有の設定はgoogle_identity_platform_default_supported_idp_configでは
  # appleSignInConfig（teamId, keyId, privateKey）を直接管理できないため、
  # 手動設定またはFirebase REST APIでの管理が必要
}
