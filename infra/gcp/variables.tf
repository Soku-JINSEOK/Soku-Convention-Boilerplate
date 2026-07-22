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

variable "wif_pool_id" {
  description = "Workload Identity Pool ID."
  type        = string
  default     = "github-actions"
}

variable "wif_provider_id" {
  description = "Workload Identity Provider ID under the pool."
  type        = string
  default     = "gha"
}

variable "enabled_apis" {
  description = "GCP APIs that must be enabled for this stack."
  type        = list(string)
  default = [
    "artifactregistry.googleapis.com",
    "cloudrun.googleapis.com",
    "iam.googleapis.com",
    "sts.googleapis.com",
    "compute.googleapis.com",
  ]
}
