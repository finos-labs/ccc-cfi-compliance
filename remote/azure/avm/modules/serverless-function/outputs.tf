output "name" {
  description = "The name of the function app."
  value       = module.avm_res_web_site.name
}

output "resource_id" {
  description = "The resource ID of the function app."
  value       = nonsensitive(module.avm_res_web_site.resource_id)
}
