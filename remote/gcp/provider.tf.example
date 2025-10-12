# GCP Provider with default labels
# Note: GCP uses "labels" instead of "tags"
# Version constraint is intentionally omitted - let the module specify its required version
# project_id will be set via TF_VAR_project_id environment variable

provider "google" {
  # Default labels applied to ALL GCP resources that support labels
  default_labels = {
    environment      = "cfi-test"
    managed_by       = "terraform"
    project          = "ccc-cfi-compliance"
    auto_cleanup     = "true"
    cfi_target_id    = var.target_id_sanitized
    github_workflow  = "cfi-build"
    github_run_id    = var.github_run_id_sanitized
  }
}

# Variables for CFI testing metadata
# Note: project_id and region are typically declared by the module itself
# Only declare CFI-specific variables here

variable "target_id" {
  description = "CFI Target ID (e.g., gcp-storage-bucket)"
  type        = string
  default     = "local-test"
}

# GCP labels must be lowercase and can only contain lowercase letters, numbers, hyphens, and underscores
variable "target_id_sanitized" {
  description = "Sanitized CFI Target ID for GCP labels (lowercase, hyphens)"
  type        = string
  default     = "local-test"
}

variable "github_run_id" {
  description = "GitHub Actions run ID"
  type        = string
  default     = "local"
}

variable "github_run_id_sanitized" {
  description = "Sanitized GitHub run ID for GCP labels"
  type        = string
  default     = "local"
}

# Note: GCP label requirements:
# - Keys and values must be lowercase
# - Only letters, numbers, hyphens, and underscores allowed
# - Maximum 63 characters per key/value
# - No spaces allowed (use hyphens or underscores instead)

