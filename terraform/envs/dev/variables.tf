variable "project_id" {
  description = "GCP Project ID"
  type        = string
  default     = "api-project-732262258565"
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "asia-northeast1"
}

variable "container_image" {
  description = "Full container image URL (set by CI/CD)"
  type        = string
  default     = "asia-northeast1-docker.pkg.dev/api-project-732262258565/kyouen-repo/kyouen-server:latest"
}

variable "service_account_email" {
  description = "Service account email for Cloud Run"
  type        = string
  default     = "732262258565-compute@developer.gserviceaccount.com"
}

# 機密情報（Secret Manager または CI/CD 経由で設定）
variable "twitter_client_id" {
  description = "Twitter OAuth Client ID"
  type        = string
  sensitive   = true
}

variable "twitter_client_secret" {
  description = "Twitter OAuth Client Secret"
  type        = string
  sensitive   = true
}

variable "apple_client_id" {
  description = "Apple Services ID"
  type        = string
  sensitive   = true
}

variable "apple_client_secret" {
  description = "Apple client secret JWT"
  type        = string
  sensitive   = true
}
