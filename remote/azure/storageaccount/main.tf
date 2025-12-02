variable "storage_account_name" {
  description = "Azure storage account name (set via TF_VAR_storage_account_name or AZURE_STORAGE_ACCOUNT)"
  type        = string
  default     = "storagecfitesting2025"
}

# Storage account for compliance testing
module "storage_account" {
  source = "git::https://github.com/Azure/terraform-azurerm-avm-res-storage-storageaccount.git?ref=main"
  location = "eastus"
  resource_group_name = "cfi_test"
  name = var.storage_account_name

  account_tier             = "Standard"
  account_replication_type = "LRS"  # Locally redundant - cheapest option
  access_tier              = "Hot"  # Hot is default, but explicit

  # Enable versioning required for immutability policies
  is_hns_enabled = false
  blob_properties = {
    versioning_enabled = true
    
    # Container delete retention for soft delete (CN03 tests)
    container_delete_retention_policy = {
      days = 3
    }
    
    # Blob delete retention for soft delete (CN03 tests)
    delete_retention_policy = {
      days = 3
    }
  }

  # Create default container with immutability policy (CN04 tests - 3 day retention)
  containers = {
    ccc-test-container = {
      name                  = "ccc-test-container"
      container_access_type = "private"
      
      # Time-based retention policy for WORM compliance
      immutability_policy = {
        immutability_period_in_days = 3
        policy_mode                  = "Unlocked"  # Unlocked allows testing, Locked is for production
      }
    }
  }
}
