variable "project_id" {
  description = "GCP project ID"
  type        = string
  default     = "dfloo-profile"
}

variable "region" {
  description = "GCP region for resources"
  type        = string
  default     = "us-west1"
}

variable "cluster_name" {
  description = "GKE cluster name"
  type        = string
  default     = "dfloo-profile-cluster"
}

variable "enable_autopilot" {
  description = "Whether to enable GKE Autopilot"
  type        = bool
  default     = true
}

variable "k8s_namespace" {
  description = "Kubernetes namespace for workload identity bindings"
  type        = string
  default     = "dfloo-profile"
}

variable "k8s_service_account_name" {
  description = "Google service account name used by the KSA"
  type        = string
  default     = "dfloo-profile-k8s"
}

variable "cloudbuild_service_account_name" {
  description = "Google service account name used by Cloudbuild"
  type        = string
  default     = "dfloo-profile-builder"
}

variable "required_apis" {
  description = "APIs required for the project"
  type        = list(string)
  default = [
    "container.googleapis.com",
    "secretmanager.googleapis.com",
    "containerregistry.googleapis.com",
    "cloudbuild.googleapis.com",
    "storage.googleapis.com",
    "iam.googleapis.com",
  ]
}
