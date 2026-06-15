output "resource_group_name" {
  description = "Name of the AVM testing resource group."
  value       = azurerm_resource_group.this.name
}

output "storage_account" {
  description = "Deployed storage account details."
  value = {
    name              = module.storage_account.name
    resource_id       = module.storage_account.resource_id
    fqdn              = module.storage_account.fqdn
    containers        = module.storage_account.containers
    default_container = module.storage_account.default_container
  }
}

output "key_vault" {
  description = "Deployed key vault details."
  value = {
    name        = module.key_vault.name
    resource_id = module.key_vault.resource_id
    uri         = module.key_vault.uri
  }
}
