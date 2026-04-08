# Azure Provider with default tags
# Tags can be set at provider level or resource group level
# Note: Version constraint is intentionally omitted - let the module specify its required version

provider "azurerm" {
  features {}
}

# Create a resource group with standard tags
# All resources within this group should inherit these tags
resource "azurerm_resource_group" "cfi_test" {
  name     = "rg-cfi-${var.instance_id}-${random_id.suffix.hex}"
  location = var.azure_location

  tags = {
    Environment      = "cfi-test"
    ManagedBy        = "Terraform"
    Project          = "CCC-CFI-Compliance"
    AutoCleanup      = "true"
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

variable "instance_id" {
  description = "CFI Instance ID"
  type        = string
  default     = "local-test"
}

# Note: For each Azure resource, you should also add tags explicitly:
# resource "azurerm_storage_account" "example" {
#   ...
#   tags = azurerm_resource_group.cfi_test.tags
# }

