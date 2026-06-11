locals {
  storage_account_name = "avmstor${var.instance_id}"
  default_container    = "ccc-avm-test-container-${var.instance_id}"
}

# Resource group for AVM testing.
# Managed as a resource to allow creation, but we import it if it already exists
# because it is excluded from the automated cleanup (nuke).
resource "azurerm_resource_group" "this" {
  name     = "avm-testing"
  location = var.location
}

module "storage_account" {
  source = "./modules/storage-account"

  location  = var.location
  parent_id = azurerm_resource_group.this.id
  name      = local.storage_account_name

  allow_nested_items_to_be_public   = var.allow_nested_items_to_be_public
  blob_properties                   = var.blob_properties
  default_to_oauth_authentication   = var.default_to_oauth_authentication
  https_traffic_only_enabled        = var.https_traffic_only_enabled
  infrastructure_encryption_enabled = var.infrastructure_encryption_enabled
  min_tls_version                   = var.min_tls_version
  network_rules                     = var.network_rules
  public_network_access_enabled     = var.public_network_access_enabled
  shared_access_key_enabled         = var.shared_access_key_enabled
  default_container                 = local.default_container
}
