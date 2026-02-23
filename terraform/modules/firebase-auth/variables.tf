variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "firebase_subdomain" {
  description = "Firebase subdomain (same as project_id in most cases)"
  type        = string
}

variable "anonymous_auth_enabled" {
  description = "Whether to enable anonymous authentication"
  type        = bool
  default     = false
}

variable "additional_authorized_domains" {
  description = "Additional authorized domains for Firebase Auth"
  type        = list(string)
  default     = []
}

variable "twitter_idp_enabled" {
  description = "Whether to enable Twitter IdP"
  type        = bool
  default     = false
}

variable "twitter_client_id" {
  description = "Twitter OAuth Client ID (API Key)"
  type        = string
  default     = ""
  sensitive   = true
}

variable "twitter_client_secret" {
  description = "Twitter OAuth Client Secret"
  type        = string
  default     = ""
  sensitive   = true
}

variable "apple_idp_enabled" {
  description = "Whether to enable Apple IdP"
  type        = bool
  default     = false
}

variable "apple_client_id" {
  description = "Apple Services ID (Bundle ID)"
  type        = string
  default     = ""
  sensitive   = true
}

variable "apple_client_secret" {
  description = "Apple client secret (JWT)"
  type        = string
  default     = ""
  sensitive   = true
}
