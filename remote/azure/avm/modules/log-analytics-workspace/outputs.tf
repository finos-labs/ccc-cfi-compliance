output "name" {
  description = "The name of the Log Analytics workspace."
  value       = module.avm_res_operationalinsights_workspace.resource.name
}

output "resource_id" {
  description = "The resource ID of the Log Analytics workspace."
  value       = module.avm_res_operationalinsights_workspace.resource_id
}
