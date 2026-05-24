variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "database_id" {
  description = "Firestore database ID (e.g. 'kyouen')"
  type        = string
}

variable "location_id" {
  description = "Firestore database location (e.g. 'asia-northeast1')"
  type        = string
  default     = "asia-northeast1"
}
