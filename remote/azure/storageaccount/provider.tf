# Azure Provider configuration
# Subscription ID can be set via:
#   - ARM_SUBSCRIPTION_ID environment variable
#   - TF_VAR_subscription_id environment variable
#   - az account set --subscription <id>

variable "subscription_id" {
  description = "Azure subscription ID (optional if using az cli or ARM_SUBSCRIPTION_ID)"
  type        = string
  default     = null
}

provider "azurerm" {
  features {}
  
  subscription_id = var.subscription_id
  
  # Use Azure AD authentication for storage operations instead of access keys
  storage_use_azuread = true
}
