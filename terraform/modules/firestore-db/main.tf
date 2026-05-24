terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 7.0"
    }
  }
}

# 名前付き Firestore データベース（Datastore モード、東京リージョン）
resource "google_firestore_database" "kyouen" {
  project     = var.project_id
  name        = var.database_id
  location_id = var.location_id
  type        = "DATASTORE_MODE"

  deletion_policy = "DELETE"
}
