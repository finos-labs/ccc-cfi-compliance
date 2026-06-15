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
  private_endpoints                 = var.private_endpoints
  public_network_access_enabled     = var.public_network_access_enabled
  shared_access_key_enabled         = var.shared_access_key_enabled
  managed_identities                = var.managed_identities
  customer_managed_key              = var.customer_managed_key

  # Account-level blob versioning stays on via var.blob_properties
  # (CCC.ObjStor.CN04/CN05). We intentionally do NOT enable container-level
  # version-level immutability (WORM) here, for two reasons:
  #   1. It is wrong for the function's `deploymentpackage` container, which the
  #      Flex Consumption runtime must be able to overwrite.
  #   2. On the test data container an *unlocked* immutability flag does not
  #      satisfy CCC.ObjStor.CN03.AR02 (the retention policy MUST be
  #      irrevocable), yet it still forces the container create to depend on the
  #      versioning update, which races on a cold apply
  #      ("Required feature Versioning is disabled").
  #
  # Realizing CN03 fully IS possible: on the DATA container only (never the
  # function package container), enable version-level immutability and then
  # *lock* a time-based retention policy, e.g.
  #     immutable_storage_with_versioning = { enabled = true }
  # plus a locked Microsoft.Storage immutabilityPolicy with
  # immutabilityPeriodSinceCreationInDays (and optionally a legal hold). Once
  # locked, the policy cannot be shortened or removed, so objects are
  # irrevocably retained for the period. The trade-off is that the account /
  # container then cannot be deleted until every blob's retention expires, which
  # is why it is omitted from this tear-down-able reference deployment.

  containers = merge(
    {
      (var.default_container) = {
        name          = var.default_container
        public_access = "None"
      }
    },
    {
      for key, container in var.extra_containers : key => {
        name          = container.name
        public_access = "None"
      }
    }
  )
}
