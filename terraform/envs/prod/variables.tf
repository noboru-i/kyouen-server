variable "project_id" {
  description = "GCP Project ID"
  type        = string
  default     = "my-android-server"
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "asia-northeast1"
}

variable "container_image" {
  description = "Full container image URL (set by CI/CD)"
  type        = string
  default     = "asia-northeast1-docker.pkg.dev/my-android-server/kyouen-repo/kyouen-server:latest"
}

variable "service_account_email" {
  description = "Service account email for Cloud Run"
  type        = string
  default     = "495839147593-compute@developer.gserviceaccount.com"
}
