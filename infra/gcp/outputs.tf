output "project_id" {
  description = "GCP project used by this stack."
  value       = var.project_id
}

output "region" {
  description = "Region used for Artifact Registry and Cloud Run."
  value       = var.region
}

output "service_name" {
  description = "Cloud Run service name."
  value       = var.service_name
}

output "service_url" {
  description = "Cloud Run service URL."
  value       = var.deploy_runtime ? google_cloud_run_service.service[0].status[0].url : null
}

output "runtime_service_account_email" {
  description = "Runtime service account bound to the Cloud Run service."
  value       = google_service_account.cloud_run_runtime.email
}

output "artifact_registry_repository_id" {
  description = "Artifact Registry repository id."
  value       = google_artifact_registry_repository.repository.repository_id
}

output "artifact_registry_repository_url" {
  description = "Full Artifact Registry repository name."
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.repository.repository_id}"
}

output "deployer_service_account_email" {
  description = "Service account used by deployment from GitHub Actions."
  value       = google_service_account.github_actions_deployer.email
}

output "wif_pool_id" {
  description = "Workload Identity Pool ID for deployment."
  value       = var.enable_wif ? google_iam_workload_identity_pool.github[0].workload_identity_pool_id : null
}

output "wif_provider_id" {
  description = "Workload Identity Provider ID for deployment."
  value       = var.enable_wif ? google_iam_workload_identity_pool_provider.github[0].workload_identity_pool_provider_id : null
}

output "wif_provider_name" {
  description = "Full Workload Identity Provider resource name for GitHub Actions."
  value       = var.enable_wif ? google_iam_workload_identity_pool_provider.github[0].name : null
}
