data "google_project" "current" {
  project_id = var.project_id
}

resource "google_project_service" "required_apis" {
  for_each = toset(var.enabled_apis)

  project = var.project_id
  service = each.value

  disable_dependent_services = false
}

resource "google_artifact_registry_repository" "repository" {
  location      = var.region
  repository_id = var.artifact_repository
  description   = "Container images for ${var.service_name}"
  format        = "DOCKER"

  depends_on = [google_project_service.required_apis]
}

resource "google_service_account" "cloud_run_runtime" {
  project      = var.project_id
  account_id   = "${substr(var.service_name, 0, 20)}-runtime"
  display_name = "Cloud Run runtime identity for ${var.service_name}"

  depends_on = [google_project_service.required_apis]
}

resource "google_service_account" "github_actions_deployer" {
  project      = var.project_id
  account_id   = "${substr(var.service_name, 0, 15)}-gh-deployer"
  display_name = "GitHub Actions deployer identity for ${var.service_name}"

  depends_on = [google_project_service.required_apis]
}

resource "google_project_iam_member" "deployer_run_admin" {
  project = var.project_id
  role    = "roles/run.admin"
  member  = "serviceAccount:${google_service_account.github_actions_deployer.email}"
}

resource "google_project_iam_member" "deployer_artifact_registry_writer" {
  project = var.project_id
  role    = "roles/artifactregistry.writer"
  member  = "serviceAccount:${google_service_account.github_actions_deployer.email}"
}

resource "google_project_iam_member" "deployer_service_account_user" {
  project = var.project_id
  role    = "roles/iam.serviceAccountUser"
  member  = "serviceAccount:${google_service_account.github_actions_deployer.email}"
}

resource "google_project_iam_member" "deployer_token_creator" {
  project = var.project_id
  role    = "roles/iam.serviceAccountTokenCreator"
  member  = "serviceAccount:${google_service_account.github_actions_deployer.email}"
}

resource "google_artifact_registry_repository_iam_member" "deployer_repository_writer" {
  location   = var.region
  repository = google_artifact_registry_repository.repository.repository_id
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:${google_service_account.github_actions_deployer.email}"
}

resource "google_cloud_run_service" "service" {
  count    = var.deploy_runtime ? 1 : 0
  name     = var.service_name
  location = var.region

  template {
    metadata {
      annotations = {
        "autoscaling.knative.dev/minScale" = tostring(var.min_instances)
        "autoscaling.knative.dev/maxScale" = tostring(var.max_instances)
      }
    }

    spec {
      service_account_name = google_service_account.cloud_run_runtime.email
      timeout_seconds      = 300

      containers {
        image = var.image_uri

        ports {
          container_port = var.container_port
        }

        dynamic "env" {
          for_each = var.environment_variables

          content {
            name  = env.key
            value = env.value
          }
        }
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

  depends_on = [
    google_project_service.required_apis,
    google_artifact_registry_repository.repository,
  ]

  lifecycle {
    precondition {
      condition     = !var.deploy_runtime || var.image_uri != null
      error_message = "image_uri is required when deploy_runtime is true."
    }
  }
}

resource "google_cloud_run_service_iam_member" "public_invoker" {
  count    = var.deploy_runtime && var.allow_unauthenticated ? 1 : 0
  location = google_cloud_run_service.service[0].location
  project  = var.project_id
  service  = google_cloud_run_service.service[0].name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

resource "google_iam_workload_identity_pool" "github" {
  count                     = var.enable_wif ? 1 : 0
  workload_identity_pool_id = var.wif_pool_id
  display_name              = "github-actions-${var.service_name}"
  description               = "Workload Identity Pool for GitHub Actions deploy workflows."

  depends_on = [google_project_service.required_apis]
}

resource "google_iam_workload_identity_pool_provider" "github" {
  count                              = var.enable_wif ? 1 : 0
  workload_identity_pool_id          = google_iam_workload_identity_pool.github[0].workload_identity_pool_id
  workload_identity_pool_provider_id = var.wif_provider_id
  display_name                       = "github-provider-${var.service_name}"
  description                        = "OIDC provider for GitHub Actions"
  disabled                           = false
  attribute_mapping = {
    "google.subject"       = "assertion.sub"
    "attribute.repository" = "assertion.repository"
  }
  attribute_condition = "assertion.repository == \"${var.github_org}/${var.github_repo}\""
  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}

resource "google_service_account_iam_member" "github_deployer_wi" {
  count = var.enable_wif ? 1 : 0

  service_account_id = google_service_account.github_actions_deployer.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/projects/${data.google_project.current.number}/locations/global/workloadIdentityPools/${google_iam_workload_identity_pool.github[0].workload_identity_pool_id}/attribute.repository/${var.github_org}/${var.github_repo}"
}
