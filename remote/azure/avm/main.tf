locals {
  storage_account_name = "avmstor${var.instance_id}"
  default_container    = "ccc-avm-test-container-${var.instance_id}"
  key_vault_name       = "avmkv${var.instance_id}"
}

data "azurerm_client_config" "current" {}

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

module "key_vault" {
  source = "./modules/key-vault"

  location            = var.location
  resource_group_name = azurerm_resource_group.this.name
  name                = local.key_vault_name
  tenant_id           = data.azurerm_client_config.current.tenant_id

  enabled_for_deployment          = var.enabled_for_deployment
  enabled_for_disk_encryption     = var.enabled_for_disk_encryption
  enabled_for_template_deployment = var.enabled_for_template_deployment
  legacy_access_policies_enabled  = var.legacy_access_policies_enabled
  network_acls                    = var.network_acls
  public_network_access_enabled   = var.public_network_access_enabled
  purge_protection_enabled        = var.purge_protection_enabled
  sku_name                        = var.sku_name
  soft_delete_retention_days      = var.soft_delete_retention_days
}
