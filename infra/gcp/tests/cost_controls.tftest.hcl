mock_provider "google" {}

run "cost_controls_safe_defaults" {
  command = plan

  variables {
    project_id     = "test-project-123"
    deploy_runtime = false
    enable_wif     = false
  }

  assert {
    condition     = length(google_billing_budget.project) == 0
    error_message = "Budget creation must remain opt-in."
  }

  assert {
    condition     = google_artifact_registry_repository.repository.cleanup_policy_dry_run
    error_message = "Artifact cleanup must start in dry-run mode."
  }

  assert {
    condition     = length(google_artifact_registry_repository.repository.cleanup_policies) == 4
    error_message = "The repository must define delete and retention cleanup policies."
  }
}

run "budget_and_cleanup_activation" {
  command = plan

  variables {
    project_id               = "test-project-123"
    deploy_runtime           = false
    enable_wif               = false
    enable_budget_alerts     = true
    billing_account_id       = "ABCDEF-123456-ABCDEF"
    monthly_budget_amount    = 1500
    artifact_cleanup_dry_run = false
    artifact_keep_count      = 5
  }

  assert {
    condition = (
      length(google_project_service.billing_budgets) == 1 &&
      length(google_billing_budget.project) == 1
    )
    error_message = "Budget activation must enable its API and create one budget."
  }

  assert {
    condition = (
      google_billing_budget.project[0].amount[0].specified_amount[0].currency_code == "JPY" &&
      google_billing_budget.project[0].amount[0].specified_amount[0].units == "1500" &&
      length(google_billing_budget.project[0].threshold_rules) == 3
    )
    error_message = "The budget must use the reviewed amount and three current-spend thresholds."
  }

  assert {
    condition     = !google_artifact_registry_repository.repository.cleanup_policy_dry_run
    error_message = "Explicit activation must disable Artifact cleanup dry-run."
  }
}
