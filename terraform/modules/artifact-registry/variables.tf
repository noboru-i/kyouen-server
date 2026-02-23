variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "location" {
  description = "Artifact Registry location"
  type        = string
  default     = "asia-northeast1"
}
