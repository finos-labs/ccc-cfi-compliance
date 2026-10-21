output "name" {
  description = "The name of the key vault."
  value       = module.avm_res_keyvault_vault.name
}

output "resource_id" {
  description = "The resource ID of the key vault."
  value       = module.avm_res_keyvault_vault.resource_id
}

output "uri" {
  description = "The URI of the key vault, used for data-plane access."
  value       = module.avm_res_keyvault_vault.uri
}
