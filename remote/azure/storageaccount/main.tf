variable "storage_account_name" {
  description = "Azure storage account name"
  type        = string
  default     = "storagecfitesting2026"
}

# Resource group for CFI testing
resource "azurerm_resource_group" "cfi_test" {
  name     = "cfi_test"
  location = "eastus"
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

# Create the resource group if it doesn't exist
resource "azurerm_resource_group" "this" {
  name     = var.resource_group_name
  location = var.location
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

  # Create default container with immutability policy (CN04 tests - 3 day retention)
  containers = {
    ccc-test-container = {
      name                  = "ccc-test-container"
      container_access_type = "private"
      
      # Time-based retention policy for WORM compliance (CCC.ObjStor.CN03.AR02)
      immutability_policy = {
        immutability_period_in_days = 3
        policy_mode                  = "Locked"   # Required for policy; irreversible until period expires
      }
    }
  }
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

