terraform {
  required_version = ">= 1.5"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
  }
}

variable "project_id" {
  type    = string
  default = "task-tracker-app-1-a4a61"
}

# Closest Cloud Run region to the Neon project (aws-us-east-2).
variable "region" {
  type    = string
  default = "us-east4"
}

# Set with TF_VAR_database_url="$(terraform -chdir=../neon output -raw database_url)". It lands in this
# directory's local state file (git-ignored), so keep that file private.
variable "database_url" {
  type      = string
  sensitive = true
}

# The first apply runs Google's hello image; CI replaces it with the real one, and ignore_changes keeps
# Terraform from reverting that.
variable "image" {
  type    = string
  default = "us-docker.pkg.dev/cloudrun/container/hello"
}

provider "google" {
  project = var.project_id
  region  = var.region
}

resource "google_project_service" "apis" {
  for_each           = toset(["run.googleapis.com", "artifactregistry.googleapis.com", "secretmanager.googleapis.com"])
  service            = each.key
  disable_on_destroy = false
}

resource "google_artifact_registry_repository" "app" {
  repository_id = "user-auth"
  location      = var.region
  format        = "DOCKER"
  depends_on    = [google_project_service.apis]
}

resource "google_service_account" "runtime" {
  account_id   = "user-auth-runtime"
  display_name = "user_auth Cloud Run runtime"
}

resource "google_secret_manager_secret" "database_url" {
  secret_id = "user-auth-database-url"
  replication {
    auto {}
  }
  depends_on = [google_project_service.apis]
}

resource "google_secret_manager_secret_version" "database_url" {
  secret      = google_secret_manager_secret.database_url.id
  secret_data = var.database_url
}

resource "google_secret_manager_secret_iam_member" "runtime" {
  secret_id = google_secret_manager_secret.database_url.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.runtime.email}"
}

resource "google_cloud_run_v2_service" "app" {
  name                = "user-auth"
  location            = var.region
  ingress             = "INGRESS_TRAFFIC_ALL"
  deletion_protection = false

  template {
    service_account = google_service_account.runtime.email

    # Scale to zero for the free tier; the per-IP rate limiter is per instance, so cap instances low.
    scaling {
      min_instance_count = 0
      max_instance_count = 2
    }

    containers {
      image = var.image

      resources {
        limits   = { cpu = "1", memory = "512Mi" }
        cpu_idle = true
      }

      env {
        name  = "AUTH_COOKIE_SECURE"
        value = "true"
      }
      env {
        name = "AUTH_DATABASE_URL"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.database_url.secret_id
            version = "latest"
          }
        }
      }
    }
  }

  lifecycle {
    ignore_changes = [template[0].containers[0].image, client, client_version]
  }

  depends_on = [
    google_secret_manager_secret_version.database_url,
    google_secret_manager_secret_iam_member.runtime,
  ]
}

# Public demo: anyone may call the service.
resource "google_cloud_run_v2_service_iam_member" "public" {
  name     = google_cloud_run_v2_service.app.name
  location = var.region
  role     = "roles/run.invoker"
  member   = "allUsers"
}

output "service_url" {
  value = google_cloud_run_v2_service.app.uri
}

output "image_repository" {
  value = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.app.repository_id}/app"
}
