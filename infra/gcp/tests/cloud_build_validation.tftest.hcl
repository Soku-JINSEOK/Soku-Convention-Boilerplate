mock_provider "google" {}

run "cloud_build_validation_disabled" {
  command = plan

  variables {
    project_id     = "test-project-123"
    deploy_runtime = false
    enable_wif     = false
    github_org     = "Soku-JINSEOK"
    github_repo    = "Soku-Convention-Boilerplate"
  }

  assert {
    condition = (
      length(google_project_service.cloud_build) == 0 &&
      length(google_service_account.cloud_build_validation) == 0 &&
      length(google_project_iam_member.cloud_build_validation_log_writer) == 0 &&
      length(google_cloudbuild_trigger.pull_request) == 0 &&
      length(google_cloudbuild_trigger.main) == 0
    )
    error_message = "Cloud Build validation resources must be absent by default."
  }
}

run "cloud_build_validation_enabled" {
  command = plan

  variables {
    project_id                    = "test-project-123"
    deploy_runtime                = false
    enable_wif                    = false
    enable_cloud_build_validation = true
    github_org                    = "Soku-JINSEOK"
    github_repo                   = "Soku-Convention-Boilerplate"
  }

  assert {
    condition = (
      length(google_project_service.cloud_build) == 1 &&
      length(google_service_account.cloud_build_validation) == 1 &&
      length(google_project_iam_member.cloud_build_validation_log_writer) == 1 &&
      length(google_cloudbuild_trigger.pull_request) == 1 &&
      length(google_cloudbuild_trigger.main) == 1
    )
    error_message = "Enabling Cloud Build validation must create one identity, one IAM binding, and exactly two triggers."
  }

  assert {
    condition     = google_project_service.cloud_build[0].service == "cloudbuild.googleapis.com"
    error_message = "Enabling validation must enable the Cloud Build API."
  }

  assert {
    condition = (
      startswith(google_service_account.cloud_build_validation[0].account_id, "cb-") &&
      endswith(google_service_account.cloud_build_validation[0].account_id, "-ci") &&
      length(google_service_account.cloud_build_validation[0].account_id) <= 30
    )
    error_message = "The validation service-account ID must preserve its prefix and suffix within the GCP length limit."
  }

  assert {
    condition = (
      google_cloudbuild_trigger.pull_request[0].location == "asia-northeast1" &&
      google_cloudbuild_trigger.pull_request[0].name == "soku-convention-boilerplate-pr" &&
      google_cloudbuild_trigger.pull_request[0].filename == "cloudbuild/validation.yaml" &&
      google_cloudbuild_trigger.pull_request[0].github[0].owner == "Soku-JINSEOK" &&
      google_cloudbuild_trigger.pull_request[0].github[0].name == "Soku-Convention-Boilerplate" &&
      google_cloudbuild_trigger.pull_request[0].github[0].pull_request[0].branch == "^main$" &&
      google_cloudbuild_trigger.pull_request[0].github[0].pull_request[0].comment_control == "COMMENTS_ENABLED"
    )
    error_message = "The pull-request trigger must use the reviewed first-generation GitHub policy."
  }

  assert {
    condition = (
      google_cloudbuild_trigger.main[0].location == "asia-northeast1" &&
      google_cloudbuild_trigger.main[0].name == "soku-convention-boilerplate-main" &&
      google_cloudbuild_trigger.main[0].filename == "cloudbuild/validation.yaml" &&
      google_cloudbuild_trigger.main[0].github[0].push[0].branch == "^main$" &&
      toset(google_cloudbuild_trigger.main[0].included_files) == toset([
        ".github/cloudbuild-validation.test.mjs",
        ".github/deploy-gcp.test.mjs",
        "cloudbuild/**",
        "infra/gcp/**",
        "scripts/gcp-bootstrap.sh",
        "templates/gcloud/**",
      ])
    )
    error_message = "The main trigger must validate only GCP-related pushes to main in Tokyo."
  }
}
