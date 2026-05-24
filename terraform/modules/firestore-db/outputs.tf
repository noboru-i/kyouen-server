output "database_id" {
  description = "Firestore database ID"
  value       = google_firestore_database.kyouen.name
}
