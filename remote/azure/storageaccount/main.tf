variable "storage_account_name" {
  description = "Azure storage account name"
  type        = string
  default     = "storagecfitesting2026"
}

variable "location" {
  description = "Azure region"
  type        = string
  default     = "eastus"
}

variable "resource_group_name" {
  description = "Azure resource group name"
  type        = string
  default     = "cfi_test"
}

# Resource group for CFI testing
# Managed as a resource to allow creation, but we import it if it already exists
# because it is excluded from the automated cleanup (nuke).
resource "azurerm_resource_group" "this" {
  name     = var.resource_group_name
  location = var.location

  tags = {
    CCC_INFRA_DONT_DELETE = "true"  # Excluded from nuke (foundation)
  }
}


# Log Analytics workspace for Azure Monitor diagnostics (CN09.AR01)
# Azure Policy/Defender may auto-create blob-diagnostic-setting targeting this workspace
resource "azurerm_log_analytics_workspace" "storage_diag" {
  name                = "cfi-storage-diag"
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
  name                = var.storage_account_name

  tags = {
    CCC_INFRA_DONT_DELETE = "true"  # Excluded from nuke (immutability policy)
  }

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
    ccc-test-container-2 = {
      name     = "ccc-test-container-2"
      public_access = "None"
      immutable_storage_with_versioning = {
        enabled = true  # Required before immutability policy can be set
      }
      tags = {
        CCC_INFRA_DONT_DELETE = "true"  # Excluded from nuke
      }
    }
  }
}

# Container-level immutability policy (CN04.AR02 - object retention enforcement)
# Must be separate from module; AVM module does not support immutability_policy on containers.
resource "azurerm_storage_container_immutability_policy" "ccc_test_container_2" {
  storage_container_resource_manager_id = module.storage_account.containers["ccc-test-container-2"].id
  immutability_period_in_days            = 2
  locked                                 = true
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

