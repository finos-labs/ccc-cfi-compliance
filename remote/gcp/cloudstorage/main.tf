variable "gcp_project_id" {
  type = string
}

variable "instance_id" {
  type = string
}

module "cloud_storage" {
  source  = "terraform-google-modules/cloud-storage/google//modules/simple_bucket"
  version = "~> 11.0"

  name       = "ccc-test-bucket-${var.instance_id}"
  project_id = var.gcp_project_id
  location   = "us-central1"

  autoclass = true
  retention_policy = {
    retention_period = 172800  # 2 days in seconds, passing CN04.AR02
  }
}
