# Subscription ID can be set via:
#   - ARM_SUBSCRIPTION_ID environment variable
#   - TF_VAR_subscription_id environment variable
#   - az account set --subscription <id>

provider "azurerm" {
  features {}

  subscription_id = var.subscription_id

  # Align with AVM guidance: prefer Entra ID for storage data-plane access.
  storage_use_azuread = true
}
