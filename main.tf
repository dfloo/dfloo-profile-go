terraform {
  required_version = ">= 1.4.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 7.17"
    }
  }
}

locals {
  workload_pool            = "${var.project_id}.svc.id.goog"
  workload_identity_member = format("serviceAccount:%s.svc.id.goog[%s/%s]", var.project_id, var.k8s_namespace, var.k8s_service_account_name)
  cloudbuild_sa_email      = format("serviceAccount:%s@cloudbuild.gserviceaccount.com", data.google_project.current.number)
}

provider "google" {
  project = var.project_id
  region  = var.region
}

data "google_project" "current" {
  project_id = var.project_id
}

resource "google_project_service" "required" {
  for_each = toset(var.required_apis)
  project  = var.project_id
  service  = each.key

  disable_on_destroy = false
}

resource "google_project_iam_member" "cloudbuild_storage_admin" {
  project = var.project_id
  role    = "roles/storage.admin"
  member  = local.cloudbuild_sa_email
}

resource "google_service_account" "ksa_gsa" {
  account_id   = var.service_account_name
  display_name = "KSA Google Service Account"
}

resource "google_service_account_iam_member" "ksa_workload_identity_binding" {
  service_account_id = google_service_account.ksa_gsa.name
  role               = "roles/iam.workloadIdentityUser"
  member             = local.workload_identity_member
}

resource "google_project_iam_member" "secret_accessor_iam" {
  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.ksa_gsa.email}"
}

resource "google_project_iam_member" "cloudbuild_container_admin" {
  project = var.project_id
  role    = "roles/container.admin"
  member  = local.cloudbuild_sa_email
}

resource "google_container_cluster" "autopilot_cluster" {
  name             = var.cluster_name
  location         = var.region
  project          = var.project_id
  enable_autopilot = var.enable_autopilot

  workload_identity_config {
    workload_pool = local.workload_pool
  }

  secret_manager_config {
    enabled = true
  }

  depends_on = [google_project_service.required]
}
