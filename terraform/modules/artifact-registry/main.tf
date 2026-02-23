terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 7.0"
    }
  }
}

# Artifact Registry: Dockerリポジトリ
resource "google_artifact_registry_repository" "kyouen_repo" {
  project       = var.project_id
  location      = var.location
  repository_id = "kyouen-repo"
  description   = "Docker repository for kyouen-server"
  format        = "DOCKER"
  mode          = "STANDARD_REPOSITORY"
}
