terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 7.0"
    }
  }
}

# Cloud Run サービス
resource "google_cloud_run_v2_service" "kyouen_server" {
  project  = var.project_id
  name     = "kyouen-server-${var.environment}"
  location = var.region

  deletion_protection = false

  template {
    service_account = var.service_account_email

    scaling {
      max_instance_count = var.max_instances
      min_instance_count = var.min_instances
    }

    max_instance_request_concurrency = 80
    timeout                          = "300s"

    containers {
      image = var.container_image

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
        startup_cpu_boost = true
      }

      env {
        name  = "GOOGLE_CLOUD_PROJECT"
        value = var.project_id
      }

      env {
        name  = "ENVIRONMENT"
        value = var.environment
      }

      ports {
        name           = "http1"
        container_port = 8080
      }
    }
  }

  ingress = "INGRESS_TRAFFIC_ALL"

  lifecycle {
    # CI/CD (GitHub Actions) が image を管理するため、Terraform では無視する
    ignore_changes = [
      template[0].containers[0].image,
      client,
      client_version,
    ]
  }
}

# 未認証アクセスを許可（パブリックAPI）
resource "google_cloud_run_v2_service_iam_member" "public_access" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.kyouen_server.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
