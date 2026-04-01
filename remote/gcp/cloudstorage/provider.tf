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
  }
}

# Variables for CFI testing metadata
# Note: project_id and region are typically declared by the module itself
# Only declare CFI-specific variables here

# - No spaces allowed (use hyphens or underscores instead)

