output "cloud_run_url" {
  description = "Cloud Run service URL"
  value       = module.kyouen_app.cloud_run_url
}

output "artifact_registry_url" {
  description = "Artifact Registry URL"
  value       = module.kyouen_app.artifact_registry_url
}
