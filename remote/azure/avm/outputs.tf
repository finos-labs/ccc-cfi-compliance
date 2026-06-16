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

output "log_analytics_workspace" {
  description = "Deployed Log Analytics workspace details."
  value = {
    name        = module.log_analytics_workspace.name
    resource_id = module.log_analytics_workspace.resource_id
  }
}

output "virtual_network" {
  description = "Deployed support virtual network details."
  value = {
    name                         = module.virtual_network.name
    resource_id                  = module.virtual_network.resource_id
    functions_subnet_resource_id = module.virtual_network.functions_subnet_resource_id
    vm_subnet_resource_id        = module.virtual_network.vm_subnet_resource_id
    pe_subnet_resource_id        = module.virtual_network.pe_subnet_resource_id
  }
}

output "serverless_function" {
  description = "Deployed serverless function app details."
  value = {
    name        = module.serverless_function.name
    resource_id = module.serverless_function.resource_id
  }
}

output "virtual_machine" {
  description = "Deployed virtual machine details."
  value = {
    name        = module.virtual_machine.name
    resource_id = module.virtual_machine.resource_id
  }
}
