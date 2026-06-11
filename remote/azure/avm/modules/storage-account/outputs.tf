output "name" {
  description = "The name of the storage account."
  value       = module.avm_res_storage_storageaccount.name
}

output "resource_id" {
  description = "The resource ID of the storage account."
  value       = module.avm_res_storage_storageaccount.resource_id
}

output "fqdn" {
  description = "FQDNs for storage services."
  value       = module.avm_res_storage_storageaccount.fqdn
}

output "default_container" {
  description = "Name of the default test container."
  value       = var.default_container
}

output "containers" {
  description = "Map of storage containers created by the module."
  value       = module.avm_res_storage_storageaccount.containers
}
