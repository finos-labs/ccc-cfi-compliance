output "name" {
  description = "The name of the managed cluster."
  value       = module.avm_res_containerservice_managedcluster.name
}

output "resource_id" {
  description = "The resource ID of the managed cluster."
  value       = module.avm_res_containerservice_managedcluster.resource_id
}

output "private_fqdn" {
  description = "Private API-server FQDN."
  value       = module.avm_res_containerservice_managedcluster.private_fqdn
}

output "oidc_issuer_url" {
  description = "OIDC issuer URL for workload federated credentials."
  value       = module.avm_res_containerservice_managedcluster.oidc_issuer_profile_issuer_url
}
