output "name" {
  description = "The name of the virtual machine."
  value       = module.avm_res_compute_virtualmachine.name
}

output "resource_id" {
  description = "The resource ID of the virtual machine."
  value       = nonsensitive(module.avm_res_compute_virtualmachine.resource_id)
}
