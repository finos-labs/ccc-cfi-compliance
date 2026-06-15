module "avm_res_keyvault_vault" {
  source  = "Azure/avm-res-keyvault-vault/azurerm"
  version = "0.10.2"

  location            = var.location
  name                = var.name
  resource_group_name = var.resource_group_name
  tenant_id           = var.tenant_id

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
