terraform {
  required_version = ">= 1.15.3, < 1.16.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "= 6.50.0"
    }
  }

  backend "gcs" {}
}

provider "google" {
  project = var.project_id
  region  = var.location
}
