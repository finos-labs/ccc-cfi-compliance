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

output "pe_subnet_resource_id" {
  description = "The resource ID of the private-endpoints subnet."
  value       = module.avm_res_network_virtualnetwork.subnets["pe"].resource_id
}

output "aks_subnet_resource_id" {
  description = "The resource ID of the AKS node subnet."
  value       = module.avm_res_network_virtualnetwork.subnets["aks"].resource_id
}

output "aks_api_subnet_resource_id" {
  description = "The resource ID of the AKS API-server VNet-integration subnet."
  value       = module.avm_res_network_virtualnetwork.subnets["aks_api"].resource_id
}
