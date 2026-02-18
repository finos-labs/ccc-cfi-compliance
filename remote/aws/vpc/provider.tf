# AWS Provider with default tags
# All resources created will automatically inherit these tags.
# Common tags are declared once via `owner` + `common_tags`.
# Note: Version constraint is intentionally omitted - let the module specify its required version

locals {
  common_resource_tags = merge(
    {
      Owner            = var.owner
      team             = var.team
      Environment      = "cfi-test"
      ManagedBy        = "Terraform"
      Project          = "CCC-CFI-Compliance"
      AutoCleanup      = "true"
      CFITargetID      = var.target_id
      GitHubWorkflow   = "CFI-Build"
      GitHubRunID      = var.github_run_id
      GitHubRepository = var.github_repository
    },
    var.common_tags
  )
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = local.common_resource_tags
  }
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "target_id" {
  description = "CFI Target ID (e.g., aws-vpc)"
  type        = string
  default     = "local-test"
}

variable "owner" {
  description = "Owner tag value applied to all resources."
  type        = string
  default     = "cfi-owner"
}

variable "team" {
  description = "Team tag value applied to all resources."
  type        = string
  default     = "cfi-team"
}

variable "common_tags" {
  description = "Additional common tags merged into all resources."
  type        = map(string)
  default     = {}
}

variable "github_run_id" {
  description = "GitHub Actions run ID"
  type        = string
  default     = "local"
}

variable "github_repository" {
  description = "GitHub repository"
  type        = string
  default     = "local"
}
