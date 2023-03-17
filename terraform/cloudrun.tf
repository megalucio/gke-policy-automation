/**
 * Copyright 2022 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

resource "google_service_account" "sa" {
  project      = data.google_project.project.project_id
  account_id   = "gke-policy-automation"
  display_name = "Service Account for GKE Policy Automation Cloud Run Service"
}

resource "google_artifact_registry_repository" "mirror" {
  project       = data.google_project.project.project_id
  location      = var.region
  repository_id = "gke-policy-automation"
  description   = "Repository for mirroring GKE policy automation image"
  format        = "docker"

  depends_on = [
    google_project_service.project
  ]
}

resource "google_project_iam_member" "run_invoker" {
  project = data.google_project.project.project_id
  role    = "roles/run.invoker"
  member  = "serviceAccount:${google_service_account.sa.email}"
}

resource "google_cloud_run_v2_job" "gke-policy-automation" {
  project      = data.google_project.project.project_id
  name         = var.job_name
  location     = var.region
  launch_stage = "BETA"

  template {
    template {
      service_account = google_service_account.sa.email
      containers {
        image   = "${var.region}-docker.pkg.dev/${data.google_project.project.project_id}/${google_artifact_registry_repository.mirror.name}/gke-policy-automation:latest"
        command = ["/gke-policy", "check"]
        args    = ["-c", "/etc/secrets/config.yaml"]
        env {
          name  = "GKE_POLICY_LOG"
          value = "INFO"
        }
        volume_mounts {
          name       = "configuration"
          mount_path = "/etc/secrets"
        }
      }
      volumes {
        name = "configuration"
        secret {
          secret = google_secret_manager_secret.config.id
          items {
            path    = "config.yaml"
            version = "latest"
            mode    = 292 #0444
          }
        }
      }
    }
  }
  depends_on = [
    google_project_service.project
  ]
}
