output "name" {
  description = "The name of the app service plan."
  value       = module.avm_res_web_serverfarm.name
}

output "resource_id" {
  description = "The resource ID of the app service plan."
  value       = module.avm_res_web_serverfarm.resource_id
}
