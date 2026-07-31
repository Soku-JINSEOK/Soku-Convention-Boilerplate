resource "google_logging_project_bucket_config" "cloud_build_validation" {
  project        = var.project_id
  location       = var.location
  bucket_id      = "cloud-build-validation"
  retention_days = 30
}

resource "google_logging_project_sink" "cloud_build_validation" {
  project = var.project_id
  name    = "cloud-build-validation"
  destination = format(
    "logging.googleapis.com/projects/%s/locations/%s/buckets/%s",
    var.project_id,
    google_logging_project_bucket_config.cloud_build_validation.location,
    google_logging_project_bucket_config.cloud_build_validation.bucket_id,
  )
  filter                 = "resource.type=\"build\""
  unique_writer_identity = false
}

resource "google_logging_project_exclusion" "default_disabled" {
  project     = var.project_id
  name        = "_Default"
  description = "Reserved rollout control; disabled until separately approved."
  filter      = "resource.type=\"build\""
  disabled    = true
}
