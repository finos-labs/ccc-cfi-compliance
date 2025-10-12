# Azure Provider with default tags
# Tags can be set at provider level or resource group level
# Note: Version constraint is intentionally omitted - let the module specify its required version

provider "azurerm" {
  features {}
}

# Create a resource group with standard tags
# All resources within this group should inherit these tags
resource "azurerm_resource_group" "cfi_test" {
  name     = "rg-cfi-${var.target_id}-${random_id.suffix.hex}"
  location = var.azure_location

  tags = {
    Environment      = "cfi-test"
    ManagedBy        = "Terraform"
    Project          = "CCC-CFI-Compliance"
    AutoCleanup      = "true"
    CFITargetID      = var.target_id
    GitHubWorkflow   = "CFI-Build"
    GitHubRunID      = var.github_run_id
    GitHubRepository = var.github_repository
  }
}

# Random suffix to avoid naming conflicts between parallel runs
resource "random_id" "suffix" {
  byte_length = 4
}

# Variables
variable "azure_location" {
  description = "Azure location"
  type        = string
  default     = "eastus"
}

variable "target_id" {
  description = "CFI Target ID (e.g., azure-storage-account)"
  type        = string
  default     = "local-test"
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

# Note: For each Azure resource, you should also add tags explicitly:
# resource "azurerm_storage_account" "example" {
#   ...
#   tags = azurerm_resource_group.cfi_test.tags
# }

