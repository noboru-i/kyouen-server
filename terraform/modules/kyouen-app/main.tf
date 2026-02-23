module "firebase_auth" {
  source = "../firebase-auth"

  project_id         = var.project_id
  firebase_subdomain = var.project_id

  anonymous_auth_enabled = false

  twitter_idp_enabled   = true
  twitter_client_id     = var.twitter_client_id
  twitter_client_secret = var.twitter_client_secret

  apple_idp_enabled   = true
  apple_client_id     = var.apple_client_id
  apple_client_secret = var.apple_client_secret
}

module "artifact_registry" {
  source = "../artifact-registry"

  project_id = var.project_id
  location   = var.region
}

module "cloud_run" {
  source = "../cloud-run"

  project_id            = var.project_id
  environment           = var.environment
  region                = var.region
  container_image       = var.container_image
  service_account_email = var.service_account_email
  max_instances         = var.max_instances
  min_instances         = var.min_instances

  depends_on = [module.artifact_registry]
}
