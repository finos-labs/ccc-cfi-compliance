# Azure Provider configuration
# Uses ARM_SUBSCRIPTION_ID environment variable set by GitHub Actions
provider "azurerm" {
  features {}
  
  # Use Azure AD authentication for storage operations instead of access keys
  storage_use_azuread = true
}
