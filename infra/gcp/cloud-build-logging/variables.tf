variable "project_id" {
  description = "GCP project ID that owns the validation log resources."
  type        = string
}

variable "location" {
  description = "Region for the Cloud Build validation log bucket."
  type        = string
  default     = "asia-northeast1"
}
