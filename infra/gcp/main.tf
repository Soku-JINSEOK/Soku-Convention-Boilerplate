data "google_project" "current" {
  project_id = var.project_id
}

locals {
  cloud_build_validation_account_id = format(
    "cb-%s-ci",
    substr(var.service_name, 0, min(24, length(var.service_name))),
  )
}

resource "google_project_service" "required_apis" {
  for_each = toset(var.enabled_apis)

  project = var.project_id
  service = each.value

  disable_dependent_services = false
}

resource "google_project_service" "cloud_build" {
  count = var.enable_cloud_build_validation ? 1 : 0

  project = var.project_id
  service = "cloudbuild.googleapis.com"

  disable_dependent_services = false
  disable_on_destroy         = false
}

resource "google_project_service" "billing_budgets" {
  count = var.enable_budget_alerts ? 1 : 0

  project = var.project_id
  service = "billingbudgets.googleapis.com"

  disable_dependent_services = false
  disable_on_destroy         = false
}

resource "google_artifact_registry_repository" "repository" {
  location               = var.region
  repository_id          = var.artifact_repository
  description            = "Container images for ${var.service_name}"
  format                 = "DOCKER"
  cleanup_policy_dry_run = var.artifact_cleanup_dry_run

  cleanup_policies {
    id     = "delete-untagged"
    action = "DELETE"

    condition {
      tag_state  = "UNTAGGED"
      older_than = "${var.artifact_untagged_retention_days * 86400}s"
    }
  }

  cleanup_policies {
    id     = "delete-expired-commit-images"
    action = "DELETE"

    condition {
      tag_state    = "TAGGED"
      tag_prefixes = ["sha-", "commit-"]
      older_than   = "${var.artifact_commit_retention_days * 86400}s"
    }
  }

  cleanup_policies {
    id     = "keep-protected-tags"
    action = "KEEP"

    condition {
      tag_state    = "TAGGED"
      tag_prefixes = ["release-", "protected-"]
    }
  }

  cleanup_policies {
    id     = "keep-recent"
    action = "KEEP"

    most_recent_versions {
      keep_count = var.artifact_keep_count
    }
  }

  depends_on = [google_project_service.required_apis]
}

resource "google_billing_budget" "project" {
  count = var.enable_budget_alerts ? 1 : 0

  billing_account = var.billing_account_id
  display_name    = "${var.service_name} monthly project budget"

  budget_filter {
    projects = ["projects/${data.google_project.current.number}"]
  }

  amount {
    specified_amount {
      currency_code = var.budget_currency_code
      units         = tostring(var.monthly_budget_amount)
    }
  }

  dynamic "threshold_rules" {
    for_each = toset([0.5, 0.8, 1.0])

    content {
      threshold_percent = threshold_rules.value
      spend_basis       = "CURRENT_SPEND"
    }
  }

  lifecycle {
    precondition {
      condition     = var.billing_account_id != null
      error_message = "billing_account_id is required when budget alerts are enabled."
    }
  }

  depends_on = [google_project_service.billing_budgets]
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

resource "google_service_account" "cloud_build_validation" {
  count = var.enable_cloud_build_validation ? 1 : 0

  project      = var.project_id
  account_id   = local.cloud_build_validation_account_id
  display_name = "Cloud Build validation identity for ${var.service_name}"

  depends_on = [
    google_project_service.cloud_build,
  ]
}

resource "google_project_iam_member" "cloud_build_validation_log_writer" {
  count = var.enable_cloud_build_validation ? 1 : 0

  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.cloud_build_validation[0].email}"
}

resource "google_project_iam_member" "deployer_run_admin" {
  project = var.project_id
  role    = "roles/run.admin"
  member  = "serviceAccount:${google_service_account.github_actions_deployer.email}"
}

resource "google_service_account_iam_member" "deployer_runtime_user" {
  service_account_id = google_service_account.cloud_run_runtime.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${google_service_account.github_actions_deployer.email}"
}

resource "google_service_account_iam_member" "deployer_self_token_creator" {
  service_account_id = google_service_account.github_actions_deployer.name
  role               = "roles/iam.serviceAccountTokenCreator"
  member             = "serviceAccount:${google_service_account.github_actions_deployer.email}"
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

resource "google_cloud_run_service_iam_member" "deployer_invoker" {
  count    = var.deploy_runtime ? 1 : 0
  location = google_cloud_run_service.service[0].location
  project  = var.project_id
  service  = google_cloud_run_service.service[0].name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.github_actions_deployer.email}"
}

resource "google_iam_workload_identity_pool" "github" {
  count                     = var.enable_wif ? 1 : 0
  workload_identity_pool_id = var.wif_pool_id
  display_name              = "github-${substr(var.service_name, 0, 20)}"
  description               = "Workload Identity Pool for GitHub Actions deploy workflows."

  depends_on = [google_project_service.required_apis]
}

resource "google_iam_workload_identity_pool_provider" "github" {
  count                              = var.enable_wif ? 1 : 0
  workload_identity_pool_id          = google_iam_workload_identity_pool.github[0].workload_identity_pool_id
  workload_identity_pool_provider_id = var.wif_provider_id
  display_name                       = "github-${substr(var.service_name, 0, 20)}"
  description                        = "OIDC provider for GitHub Actions"
  disabled                           = false
  attribute_mapping = {
    "google.subject"                = "assertion.sub"
    "attribute.repository"          = "assertion.repository"
    "attribute.repository_id"       = "assertion.repository_id"
    "attribute.repository_owner_id" = "assertion.repository_owner_id"
    "attribute.ref"                 = "assertion.ref"
    "attribute.workflow_ref"        = "assertion.workflow_ref"
  }
  attribute_condition = join(" && ", [
    "assertion.repository == \"${var.github_org}/${var.github_repo}\"",
    "assertion.repository_id == \"${coalesce(var.github_repository_id, "missing")}\"",
    "assertion.repository_owner_id == \"${coalesce(var.github_repository_owner_id, "missing")}\"",
    "assertion.ref == \"refs/heads/main\"",
    "assertion.workflow_ref == \"${var.github_org}/${var.github_repo}/.github/workflows/deploy-gcp.yml@refs/heads/main\"",
  ])
  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }

  lifecycle {
    precondition {
      condition = (
        var.github_repository_id != null &&
        var.github_repository_owner_id != null
      )
      error_message = "github_repository_id and github_repository_owner_id are required when WIF is enabled."
    }
  }
}

resource "google_service_account_iam_member" "github_deployer_wi" {
  count = var.enable_wif ? 1 : 0

  service_account_id = google_service_account.github_actions_deployer.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/projects/${data.google_project.current.number}/locations/global/workloadIdentityPools/${google_iam_workload_identity_pool.github[0].workload_identity_pool_id}/attribute.repository/${var.github_org}/${var.github_repo}"
}

resource "google_cloudbuild_trigger" "pull_request" {
  count = var.enable_cloud_build_validation ? 1 : 0

  project            = var.project_id
  location           = "global"
  name               = "soku-convention-boilerplate-pr"
  description        = "Validate pull requests without publishing or deploying artifacts."
  filename           = "cloudbuild/validation.yaml"
  service_account    = google_service_account.cloud_build_validation[0].id
  include_build_logs = "INCLUDE_BUILD_LOGS_WITH_STATUS"

  substitutions = {
    _CLOUD_BUILD_SERVICE_ACCOUNT = google_service_account.cloud_build_validation[0].id
  }

  github {
    owner = var.github_org
    name  = var.github_repo

    pull_request {
      branch          = "^main$"
      comment_control = "COMMENTS_ENABLED_FOR_EXTERNAL_CONTRIBUTORS_ONLY"
    }
  }

  depends_on = [
    google_project_iam_member.cloud_build_validation_log_writer,
    google_project_service.cloud_build,
  ]
}

resource "google_cloudbuild_trigger" "main" {
  count = var.enable_cloud_build_validation ? 1 : 0

  project            = var.project_id
  location           = "global"
  name               = "soku-convention-boilerplate-main"
  description        = "Validate main without publishing or deploying artifacts."
  filename           = "cloudbuild/validation.yaml"
  service_account    = google_service_account.cloud_build_validation[0].id
  include_build_logs = "INCLUDE_BUILD_LOGS_WITH_STATUS"

  substitutions = {
    _CLOUD_BUILD_SERVICE_ACCOUNT = google_service_account.cloud_build_validation[0].id
  }

  github {
    owner = var.github_org
    name  = var.github_repo

    push {
      branch = "^main$"
    }
  }

  depends_on = [
    google_project_iam_member.cloud_build_validation_log_writer,
    google_project_service.cloud_build,
  ]
}
