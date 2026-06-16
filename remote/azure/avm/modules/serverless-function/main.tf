module "avm_res_web_site" {
  source  = "Azure/avm-res-web-site/azurerm"
  version = "0.22.0"

  location  = var.location
  name      = var.name
  parent_id = var.parent_id

  kind                     = "functionapp"
  os_type                  = "Linux"
  service_plan_resource_id = var.service_plan_resource_id

  function_app_uses_fc1 = var.function_app_uses_fc1
  fc1_runtime_name      = var.fc1_runtime_name
  fc1_runtime_version   = var.fc1_runtime_version

  client_certificate_enabled    = var.client_certificate_enabled
  client_certificate_mode       = var.client_certificate_mode
  https_only                    = var.https_only
  managed_identities            = var.managed_identities
  maximum_instance_count        = var.maximum_instance_count
  private_endpoints             = var.private_endpoints
  public_network_access_enabled = var.public_network_access_enabled
  site_config                   = var.site_config

  # Outbound VNet integration so vnet_route_all egress policy takes effect.
  virtual_network_subnet_id = var.virtual_network_subnet_id

  # Flex Consumption backing deployment container, accessed via the pre-authorised
  # user-assigned identity (no storage account keys).
  storage_container_type            = "blobContainer"
  storage_container_endpoint        = var.storage_container_endpoint
  storage_authentication_type       = "UserAssignedIdentity"
  storage_user_assigned_identity_id = var.storage_user_assigned_identity_id
}
