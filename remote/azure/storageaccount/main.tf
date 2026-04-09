variable "instance_id" {
  description = "Unique ID for this run (lowercase letters and digits only; with prefix stgcfi the full storage account name must be ≤24 chars)"
  type        = string

  validation {
    condition = length("stgcfi${var.instance_id}") <= 24 && length("stgcfi${var.instance_id}") >= 3 && can(regex("^[a-z0-9]+$", var.instance_id))
    error_message = "instance_id must be lowercase alphanumeric, and stgcfi+instance_id must be 3–24 characters for Azure storage account naming."
  }
}

variable "location" {
  description = "Azure region"
  type        = string
  default     = "eastus"
}

# Must match testing/environment.yaml → instances[main-azure].services[object-storage]
# object-storage-retention-period-days (CN04.AR02 policy length checks).
variable "object_storage_retention_period_days" {
  type        = number
  default     = 2
  description = "Container immutability retention in days; keep in sync with testing/environment.yaml main-azure."
}

# Locked policies match real WORM-style immutability for CN04 tests. Cleanup cannot remove the
# policy (or the storage account) until blob versions exit retention — see testing/scripts/
# azure-cleanup-cfi-resource-groups.sh and .github/workflows/cfi-azure-cleanup.yml (daily retries).
variable "container_immutability_locked" {
  type        = bool
  default     = true
  description = "If false, faster teardown for dev only; compliance runs should use true."
}

locals {
  # Azure storage account name: 3–24 chars, lowercase letters and numbers only.
  # Prefix must leave room for instance_id (e.g. UTC compact ~16 chars): 6 + 16 = 22 ≤ 24.
  storage_account_name = "stgcfi${var.instance_id}"
  resource_group_name  = "cfi_test_${var.instance_id}"
  default_container   = "ccc-test-container-${var.instance_id}"
}

# Resource group for CFI testing
# Managed as a resource to allow creation, but we import it if it already exists
# because it is excluded from the automated cleanup (nuke).
resource "azurerm_resource_group" "this" {
  name     = local.resource_group_name
  location = var.location
}


# Log Analytics workspace for Azure Monitor diagnostics (CN09.AR01)
# Azure Policy/Defender may auto-create blob-diagnostic-setting targeting this workspace
resource "azurerm_log_analytics_workspace" "storage_diag" {
  name                = "cfi-storage-diag-${var.instance_id}"
  location            = azurerm_resource_group.this.location
  resource_group_name = azurerm_resource_group.this.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
}

# Storage account for compliance testing
module "storage_account" {
  source = "git::https://github.com/Azure/terraform-azurerm-avm-res-storage-storageaccount.git?ref=main"
  location            = azurerm_resource_group.this.location
  resource_group_name = azurerm_resource_group.this.name
  name                = local.storage_account_name

  account_tier             = "Standard"
  account_replication_type = "GRS"  # Geo-redundant for CN08.AR01/CN08.AR02
  access_tier              = "Hot"  # Hot is default, but explicit

  # CCC compliance: TLS, HTTPS, public access
  min_tls_version                = "TLS1_2"  # Azure max; policy may expect TLS1_3 (Azure limitation)
  https_traffic_only_enabled     = true      # Require secure transfer (CN01.AR01)
  allow_nested_items_to_be_public = false    # Block public blob access (CN05.AR01)
  shared_access_key_enabled      = true     # Required for az storage logging (Storage Analytics)

  # Enable versioning required for immutability policies
  is_hns_enabled = false
  blob_properties = {
    versioning_enabled = true
    
    # Container delete retention for soft delete (CN03.AR01 - min 7 days)
    container_delete_retention_policy = {
      days = 7
    }
    
    # Blob delete retention for soft delete (CN03.AR01 - min 7 days)
    delete_retention_policy = {
      days = 7
    }
  }

  # CN09.AR01: Blob diagnostics - Azure Policy/Defender often auto-create "blob-diagnostic-setting"
  # on new storage accounts. We skip diagnostic_settings_blob to avoid 409 Conflict (same sink).
  # If not auto-created, add diagnostic_settings_blob with a dedicated workspace.

  # Create default container with immutable storage (CN04 tests - retention policy added below)
  containers = {
    (local.default_container) = {
      name     = local.default_container
      public_access = "None"
      immutable_storage_with_versioning = {
        enabled = true  # Required before immutability policy can be set
      }
    }
  }
}

# Container-level immutability policy (CN04.AR02 - object retention enforcement)
# Must be separate from module; AVM module does not support immutability_policy on containers.
resource "azurerm_storage_container_immutability_policy" "ccc_test_container" {
  storage_container_resource_manager_id = module.storage_account.containers[local.default_container].id
  immutability_period_in_days           = var.object_storage_retention_period_days
  locked                                = var.container_immutability_locked
}

# Enable read, write, delete logging for blob service (Storage Analytics)
# Requires shared_access_key_enabled = true on the storage account
resource "terraform_data" "blob_logging" {
  triggers_replace = [module.storage_account.resource_id]

  provisioner "local-exec" {
    command = <<-EOT
      az storage logging update \
        --services b \
        --log rwd \
        --retention 7 \
        --account-name ${module.storage_account.name}
    EOT
  }
}

