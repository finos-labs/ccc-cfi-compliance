output "name" {
  description = "The name of the virtual network."
  value       = module.avm_res_network_virtualnetwork.name
}

output "resource_id" {
  description = "The resource ID of the virtual network."
  value       = module.avm_res_network_virtualnetwork.resource_id
}

output "functions_subnet_resource_id" {
  description = "The resource ID of the delegated functions integration subnet."
  value       = module.avm_res_network_virtualnetwork.subnets["functions"].resource_id
}

output "vm_subnet_resource_id" {
  description = "The resource ID of the virtual machine subnet."
  value       = module.avm_res_network_virtualnetwork.subnets["vm"].resource_id
}
