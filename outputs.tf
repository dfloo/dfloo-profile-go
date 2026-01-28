output "ksa_gsa_email" {
  description = "Google service account email used by the KSA"
  value       = google_service_account.ksa_gsa.email
  sensitive   = false
}

output "workload_pool" {
  description = "GKE workload identity pool"
  value       = local.workload_pool
  sensitive   = false
}

output "workload_identity_member" {
  description = "Workload identity member string to bind (serviceAccount:<PROJECT>.svc.id.goog[namespace/ksa])"
  value       = local.workload_identity_member
  sensitive   = false
}
