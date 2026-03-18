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

resource "google_service_account" "ksa_gsa" {
  project      = var.project_id
  account_id   = var.k8s_service_account_name
  display_name = "KSA Google Service Account"
}

resource "google_service_account" "cloudbuild_gsa" {
  project      = var.project_id
  account_id   = var.cloudbuild_service_account_name
  display_name = "Cloudbuild Google Service Account"
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

resource "google_project_iam_member" "workload_identity_user" {
  project = var.project_id
  role    = "roles/iam.workloadIdentityUser"
  member  = "serviceAccount:${google_service_account.ksa_gsa.email}"
}

resource "google_project_iam_member" "kubernetes_engine_developer" {
  project = var.project_id
  role    = "roles/container.developer"
  member  = "serviceAccount:${google_service_account.ksa_gsa.email}"
}

resource "google_project_iam_member" "kubernetes_engine_admin" {
  project = var.project_id
  role    = "roles/container.admin"
  member  = "serviceAccount:${google_service_account.ksa_gsa.email}"
}

resource "google_project_iam_member" "kubernetes_engine_cluster_admin" {
  project = var.project_id
  role    = "roles/container.clusterAdmin"
  member  = "serviceAccount:${google_service_account.ksa_gsa.email}"
}

resource "google_project_iam_member" "cloud_sql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.ksa_gsa.email}"
}

resource "google_project_iam_member" "kubernetes_engine_default_node_service_account" {
  project = var.project_id
  role    = "roles/container.defaultNodeServiceAccount"
  member  = "serviceAccount:${data.google_project.current.number}-compute@developer.gserviceaccount.com"
}

resource "google_project_iam_member" "cloudbuild_artifactregistry_writer" {
  project = var.project_id
  role    = "roles/artifactregistry.writer"
  member  = "serviceAccount:${google_service_account.cloudbuild_gsa.email}"
}

resource "google_project_iam_member" "cloudbuild_storage_admin" {
  project = var.project_id
  role    = "roles/storage.admin"
  member  = "serviceAccount:${google_service_account.cloudbuild_gsa.email}"
}

resource "google_project_iam_member" "cloudbuild_log_writer" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.cloudbuild_gsa.email}"
}

resource "google_project_iam_member" "cloudbuild_secret_accessor" {
  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.cloudbuild_gsa.email}"
}

resource "google_project_iam_member" "cloudbuild_clusters_get" {
  project = var.project_id
  role    = "roles/container.clusterViewer"
  member  = "serviceAccount:${google_service_account.cloudbuild_gsa.email}"
}

resource "google_project_iam_member" "cloudbuild_container_admin" {
  project = var.project_id
  role    = "roles/container.admin"
  member  = "serviceAccount:${google_service_account.cloudbuild_gsa.email}"
}

resource "google_compute_global_address" "api_static_ip" {
  name    = "api-static-ip"
  project = var.project_id
}

resource "google_storage_bucket" "resume_cache" {
  name                        = "dfloo-profile-resume-cache"
  project                     = var.project_id
  location                    = var.region
  uniform_bucket_level_access = true
}

resource "google_storage_bucket_iam_member" "resume_cache_object_admin" {
  bucket = google_storage_bucket.resume_cache.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.ksa_gsa.email}"
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
