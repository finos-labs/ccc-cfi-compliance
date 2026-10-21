output "name" {
  description = "The name of the private DNS zone."
  value       = var.domain_name
}

output "resource_id" {
  description = "The resource ID of the private DNS zone."
  value       = module.avm_res_network_privatednszone.resource_id
}
