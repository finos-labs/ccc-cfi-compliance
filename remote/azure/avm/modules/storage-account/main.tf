variable "default_container" {
  description = "Name of the default blob container for CFI behavioural tests."
  type        = string
}

module "avm_res_storage_storageaccount" {
  source  = "Azure/avm-res-storage-storageaccount/azurerm"
  version = "0.7.2"

  location  = var.location
  name      = var.name
  parent_id = var.parent_id

  allow_nested_items_to_be_public   = var.allow_nested_items_to_be_public
  blob_properties                   = var.blob_properties
  default_to_oauth_authentication   = var.default_to_oauth_authentication
  https_traffic_only_enabled        = var.https_traffic_only_enabled
  infrastructure_encryption_enabled = var.infrastructure_encryption_enabled
  min_tls_version                   = var.min_tls_version
  network_rules                     = var.network_rules
  public_network_access_enabled     = var.public_network_access_enabled
  shared_access_key_enabled         = var.shared_access_key_enabled

  containers = {
    (var.default_container) = {
      name          = var.default_container
      public_access = "None"
      immutable_storage_with_versioning = {
        enabled = true
      }
    }
  }
}
