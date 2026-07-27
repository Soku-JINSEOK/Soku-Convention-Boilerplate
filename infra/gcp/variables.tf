variable "project_id" {
  description = "GCP project ID."
  type        = string
}

variable "region" {
  description = "Cloud Run and Artifact Registry region."
  type        = string
  default     = "asia-northeast1"
}

variable "service_name" {
  description = "Cloud Run service name."
  type        = string
  default     = "soku-convention-boilerplate"
}

variable "artifact_repository" {
  description = "Artifact Registry repository ID."
  type        = string
  default     = "cloud-run"
}

variable "image_uri" {
  description = "Immutable container digest URI. Required when deploy_runtime is true."
  type        = string
  default     = null
  nullable    = true

  validation {
    condition     = var.image_uri == null || can(regex("@sha256:[0-9a-fA-F]{64}$", var.image_uri))
    error_message = "image_uri must be an immutable repository@sha256:<64 hex> URI."
  }
}

variable "deploy_runtime" {
  description = "Whether to create the Cloud Run runtime after foundation bootstrap."
  type        = bool
  default     = false
}

variable "container_port" {
  description = "Container port exposed by the service."
  type        = number
  default     = 8080
}

variable "min_instances" {
  description = "Minimum Cloud Run instance count."
  type        = number
  default     = 0
}

variable "max_instances" {
  description = "Maximum Cloud Run instance count."
  type        = number
  default     = 3

  validation {
    condition     = var.max_instances >= 1
    error_message = "max_instances must be at least 1."
  }
}

variable "allow_unauthenticated" {
  description = "Whether the service receives public internet traffic."
  type        = bool
  default     = false
}

variable "environment_variables" {
  description = "Environment variables injected to the Cloud Run container."
  type        = map(string)
  default     = {}
}

variable "enable_wif" {
  description = "Whether to create workload identity pool/provider resources."
  type        = bool
  default     = true
}

variable "enable_cloud_build_validation" {
  description = "Whether to create validation-only Cloud Build triggers and their dedicated identity."
  type        = bool
  default     = false
}

variable "enable_budget_alerts" {
  description = "Whether to create a project-scoped Cloud Billing budget."
  type        = bool
  default     = false
}

variable "billing_account_id" {
  description = "Cloud Billing account ID used for the optional budget."
  type        = string
  default     = null
  nullable    = true
  sensitive   = true

  validation {
    condition = (
      var.billing_account_id == null ||
      can(regex("^[0-9A-F]{6}-[0-9A-F]{6}-[0-9A-F]{6}$", var.billing_account_id))
    )
    error_message = "billing_account_id must use the XXXXXX-XXXXXX-XXXXXX form."
  }
}

variable "monthly_budget_amount" {
  description = "Monthly project budget amount in the billing account currency."
  type        = number
  default     = 1500

  validation {
    condition     = var.monthly_budget_amount > 0 && floor(var.monthly_budget_amount) == var.monthly_budget_amount
    error_message = "monthly_budget_amount must be a positive whole number."
  }
}

variable "budget_currency_code" {
  description = "ISO 4217 currency code for the monthly budget."
  type        = string
  default     = "JPY"

  validation {
    condition     = can(regex("^[A-Z]{3}$", var.budget_currency_code))
    error_message = "budget_currency_code must be an uppercase ISO 4217 code."
  }
}

variable "artifact_cleanup_dry_run" {
  description = "Whether Artifact Registry cleanup policies only report candidates."
  type        = bool
  default     = true
}

variable "artifact_untagged_retention_days" {
  description = "Days to retain untagged Artifact Registry images."
  type        = number
  default     = 7

  validation {
    condition     = var.artifact_untagged_retention_days >= 1
    error_message = "artifact_untagged_retention_days must be at least 1."
  }
}

variable "artifact_commit_retention_days" {
  description = "Days to retain ordinary commit-tagged Artifact Registry images."
  type        = number
  default     = 30

  validation {
    condition     = var.artifact_commit_retention_days >= 1
    error_message = "artifact_commit_retention_days must be at least 1."
  }
}

variable "artifact_keep_count" {
  description = "Minimum number of recent image versions retained per package."
  type        = number
  default     = 5

  validation {
    condition     = var.artifact_keep_count >= 1
    error_message = "artifact_keep_count must be at least 1."
  }
}

variable "github_org" {
  description = "GitHub organization or user hosting the repository."
  type        = string
  default     = "your-org"
}

variable "github_repo" {
  description = "GitHub repository short name used by Workload Identity Federation."
  type        = string
  default     = "your-repo"
}

variable "github_repository_id" {
  description = "Immutable numeric GitHub repository ID used by the WIF trust condition."
  type        = string
  default     = null
  nullable    = true

  validation {
    condition     = var.github_repository_id == null || can(regex("^[0-9]+$", var.github_repository_id))
    error_message = "github_repository_id must be a numeric GitHub repository ID."
  }
}

variable "github_repository_owner_id" {
  description = "Immutable numeric GitHub repository owner ID used by the WIF trust condition."
  type        = string
  default     = null
  nullable    = true

  validation {
    condition     = var.github_repository_owner_id == null || can(regex("^[0-9]+$", var.github_repository_owner_id))
    error_message = "github_repository_owner_id must be a numeric GitHub owner ID."
  }
}

variable "wif_pool_id" {
  description = "Workload Identity Pool ID."
  type        = string
  default     = "github-actions"
}

variable "wif_provider_id" {
  description = "Workload Identity Provider ID under the pool."
  type        = string
  default     = "github"

  validation {
    condition     = length(var.wif_provider_id) >= 4 && length(var.wif_provider_id) <= 32
    error_message = "wif_provider_id must contain between 4 and 32 characters."
  }
}

variable "enabled_apis" {
  description = "GCP APIs that must be enabled for this stack."
  type        = list(string)
  default = [
    "artifactregistry.googleapis.com",
    "run.googleapis.com",
    "iam.googleapis.com",
    "sts.googleapis.com",
    "compute.googleapis.com",
  ]
}
