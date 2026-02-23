variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "environment" {
  description = "Environment name (dev or prod)"
  type        = string

  validation {
    condition     = contains(["dev", "prod"], var.environment)
    error_message = "Environment must be either 'dev' or 'prod'."
  }
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "asia-northeast1"
}

variable "container_image" {
  description = "Full container image URL (set by CI/CD)"
  type        = string
}

variable "service_account_email" {
  description = "Service account email for Cloud Run"
  type        = string
}

variable "max_instances" {
  description = "Maximum number of Cloud Run instances"
  type        = number
  default     = 10
}

variable "min_instances" {
  description = "Minimum number of Cloud Run instances"
  type        = number
  default     = 0
}

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
